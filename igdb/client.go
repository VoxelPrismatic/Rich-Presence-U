package igdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTokenURL = "https://id.twitch.tv/oauth2/token"
	defaultAPIURL   = "https://api.igdb.com/v4"
	searchLimit     = 20
)

// GameHit is one title from the authed IGDB games search.
type GameHit struct {
	ID       int
	Name     string
	Slug     string
	URL      string
	CoverURL string
}

// CatalogID is the games.db id for this hit.
func (h GameHit) CatalogID() string {
	if h.ID == 0 {
		return ""
	}
	return "igdb:" + strconv.Itoa(h.ID)
}

// Client talks to Twitch (token) and api.igdb.com (games).
type Client struct {
	HTTP         *http.Client
	TokenURL     string
	APIURL       string
	UserAgent    string
	mu           sync.Mutex
	clientID     string
	clientSecret string
	token        string
	tokenExp     time.Time
}

func NewClient() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 20 * time.Second},
		TokenURL:  defaultTokenURL,
		APIURL:    defaultAPIURL,
		UserAgent: "RichPresenceQt (+https://github.com/VoxelPrismatic/Rich-Presence-U)",
	}
}

func (c *Client) SetCredentials(id, secret string) {
	id = strings.TrimSpace(id)
	secret = strings.TrimSpace(secret)
	c.mu.Lock()
	defer c.mu.Unlock()
	if id != c.clientID || secret != c.clientSecret {
		c.token = ""
		c.tokenExp = time.Time{}
	}
	c.clientID = id
	c.clientSecret = secret
}

func (c *Client) Configured() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.clientID != "" && c.clientSecret != ""
}

// Ping fetches a token to verify the stored credentials.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.accessToken(ctx)
	return err
}

// SearchGames looks up titles on platform. Requires Configured credentials.
func (c *Client) SearchGames(ctx context.Context, query string, platform Platform) ([]GameHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if !c.Configured() {
		return nil, fmt.Errorf("igdb: no credentials")
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}

	body := gamesQuery(query, platform.FilterIDs())
	raw, err := c.api(ctx, token, "/games", body)
	if err != nil {
		return nil, err
	}
	var rows []gameRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("igdb games: %w", err)
	}
	out := make([]GameHit, 0, len(rows))
	for _, row := range rows {
		if row.ID == 0 || strings.TrimSpace(row.Name) == "" {
			continue
		}
		h := GameHit{
			ID:   row.ID,
			Name: row.Name,
			Slug: row.Slug,
			URL:  strings.TrimSpace(row.URL),
		}
		if h.URL == "" {
			h.URL = GamePageURL(row.Slug)
		}
		h.CoverURL = CoverImageURL(row.Cover.ImageID)
		out = append(out, h)
	}
	return out, nil
}

type gameRow struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	URL   string `json:"url"`
	Cover struct {
		ImageID string `json:"image_id"`
	} `json:"cover"`
}

func gamesQuery(query string, platforms []int) string {
	var b strings.Builder
	b.WriteString("search ")
	b.WriteString(apicalypseString(query))
	b.WriteString("; fields id,name,slug,url,cover.image_id; limit ")
	b.WriteString(strconv.Itoa(searchLimit))
	if len(platforms) > 0 {
		b.WriteString("; where platforms = (")
		for i, id := range platforms {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.Itoa(id))
		}
		b.WriteByte(')')
	}
	b.WriteByte(';')
	return b.String()
}

func apicalypseString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " ")
	return `"` + s + `"`
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	id, secret := c.clientID, c.clientSecret
	cached, exp := c.token, c.tokenExp
	c.mu.Unlock()
	if id == "" || secret == "" {
		return "", fmt.Errorf("igdb: no credentials")
	}
	if cached != "" && time.Now().Before(exp) {
		return cached, nil
	}

	tokenURL := c.TokenURL
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	form := url.Values{}
	form.Set("client_id", id)
	form.Set("client_secret", secret)
	form.Set("grant_type", "client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setUA(req)
	resp, err := c.http().Do(req)
	if err != nil {
		return "", fmt.Errorf("igdb token: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("igdb token: %s", resp.Status)
	}
	var tok tokenResp
	if err := json.Unmarshal(raw, &tok); err != nil {
		return "", fmt.Errorf("igdb token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("igdb token: empty")
	}
	ttl := time.Duration(tok.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = time.Hour
	}
	if ttl > time.Minute {
		ttl -= time.Minute
	}
	c.mu.Lock()
	c.token = tok.AccessToken
	c.tokenExp = time.Now().Add(ttl)
	c.mu.Unlock()
	return tok.AccessToken, nil
}

func (c *Client) api(ctx context.Context, token, path, body string) ([]byte, error) {
	base := c.APIURL
	if base == "" {
		base = defaultAPIURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+path, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	id := c.clientID
	c.mu.Unlock()
	req.Header.Set("Client-ID", id)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	c.setUA(req)
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, fmt.Errorf("igdb api: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("igdb api: %s", resp.Status)
	}
	return raw, nil
}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) setUA(req *http.Request) {
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
}
