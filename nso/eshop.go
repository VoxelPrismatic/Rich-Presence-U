package nso

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

const DefaultStoreURL = "https://search.nintendo-europe.com/en/select"

type euResponse struct {
	Response struct {
		Docs []euDoc `json:"docs"`
	} `json:"response"`
}

type euDoc struct {
	Title       string          `json:"title"`
	FsID        json.RawMessage `json:"fs_id"`
	ImageSq     string          `json:"image_url_sq_s"`
	Image       string          `json:"image_url"`
	Originally  string          `json:"originally_for_t"`
	Playable    []string        `json:"playable_on_txt"`
	Systems     []string        `json:"system_names_txt"`
	Nsuid       []string        `json:"nsuid_txt"`
	TitleExtras []string        `json:"title_extras_txt"`
}

func (c *Client) storeURL() string {
	if c.StoreURL != "" {
		return c.StoreURL
	}
	return DefaultStoreURL
}

func storeFQ(system System) string {
	switch system {
	case BEE:
		return `type:GAME AND (playable_on_txt:"BEE" OR playable_on_txt:"HAC")`
	case HAC:
		return `type:GAME AND playable_on_txt:"HAC"`
	case CTR:
		return `type:GAME AND playable_on_txt:"CTR"`
	case WUP:
		return `type:GAME AND playable_on_txt:"WUP"`
	default:
		return "type:GAME"
	}
}

func assetFromPlayable(playable []string, originally string) System {
	orig := System(strings.ToUpper(originally))
	if orig.Valid() {
		return orig
	}
	for _, p := range playable {
		if sys, ok := ParseSystem(p); ok {
			return sys
		}
	}
	return ""
}

func matchesSystem(want, asset System) bool {
	if !asset.Valid() {
		return false
	}
	if want == BEE {
		return asset == BEE || asset == HAC
	}
	return want == asset
}

func (d euDoc) id() string {
	if len(d.Nsuid) > 0 && d.Nsuid[0] != "" {
		return d.Nsuid[0]
	}
	id := strings.Trim(string(d.FsID), `"`)
	if id == "" {
		return ""
	}
	return "eu:" + id
}

func (d euDoc) Game(want System) (Game, bool) {
	title := strings.TrimSpace(d.Title)
	if title == "" {
		return Game{}, false
	}
	asset := assetFromPlayable(d.Playable, d.Originally)
	if !matchesSystem(want, asset) {
		return Game{}, false
	}
	id := d.id()
	if id == "" {
		id = strings.ToLower(strings.ReplaceAll(title, " ", ""))
	}
	if want == BEE && asset == HAC {
		id = HACPrefix + strings.TrimPrefix(id, HACPrefix)
	}
	cover := d.ImageSq
	if cover == "" {
		cover = d.Image
	}
	g := Game{
		ID:          id,
		AssetSystem: asset,
		Icons:       map[Region]bool{US: true, EU: true},
		Titles:      map[Region]string{US: title, EU: title},
		CoverArt:    cover,
	}
	return g, true
}

// SearchStore looks up titles on the Nintendo eShop (Europe catalog).
func (c *Client) SearchStore(ctx context.Context, query string, system System) ([]Game, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, nil
	}
	u, err := url.Parse(c.storeURL())
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("wt", "json")
	q.Set("start", "0")
	q.Set("rows", "12")
	q.Set("fq", storeFQ(system))
	u.RawQuery = q.Encode()

	body, err := c.get(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("eshop search: %w", err)
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var resp euResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	out := make([]Game, 0, len(resp.Response.Docs))
	seen := map[string]bool{}
	for _, doc := range resp.Response.Docs {
		g, ok := doc.Game(system)
		if !ok || seen[g.ID] {
			continue
		}
		seen[g.ID] = true
		out = append(out, g)
	}
	return out, nil
}

// Remember writes a selected title (including eShop hits) into games.db.
func (c *Client) Remember(system System, g Game) error {
	if err := c.upsertGame(system, g); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	cat := c.catalogs[system]
	if cat == nil {
		cat = newCatalog()
		c.catalogs[system] = cat
	}
	if i, ok := cat.ByID[g.ID]; ok {
		cat.Games[i] = g
		return nil
	}
	cat.add(g)
	return nil
}
