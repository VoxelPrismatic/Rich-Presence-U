package nso

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	DefaultStoreURL = "https://search.nintendo-europe.com/en/select"
	defaultJPStore  = "https://search.nintendo.jp/nintendo_soft/search.json"
	usAlgoliaApp    = "U3B6GR4UA3"
	usAlgoliaKey    = "a29c6927638bfd8cee23993e51e721c9"
	usAlgoliaURL    = "https://U3B6GR4UA3-dsn.algolia.net/1/indexes/store_all_products_en_us/query"
	jpImageBase     = "https://img-eshop.cdn.nintendo.net/i/"
	usImageBase     = "https://assets.nintendo.com/image/upload/q_auto/f_auto/"
)

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
	Nsuid       []string        `json:"nsuid_txt"`
	TitleExtras []string        `json:"title_extras_txt"`
}

type jpResponse struct {
	Result struct {
		Items []jpItem `json:"items"`
	} `json:"result"`
}

type jpItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Nsuid string `json:"nsuid"`
	Hard  string `json:"hard"`
	Sform string `json:"sform"`
	IURL  string `json:"iurl"`
}

type usQuery struct {
	Query       string `json:"query"`
	HitsPerPage int    `json:"hitsPerPage"`
	Page        int    `json:"page"`
}

type usResponse struct {
	Hits []usHit `json:"hits"`
}

type usHit struct {
	Title              string `json:"title"`
	Nsuid              string `json:"nsuid"`
	Platform           string `json:"platform"`
	PlatformCode       string `json:"platformCode"`
	ProductImageSquare string `json:"productImageSquare"`
	ProductImage       string `json:"productImage"`
	DLCType            string `json:"dlcType"`
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

func applyIDPrefix(id string, want, asset System) string {
	if want == BEE && asset == HAC {
		return HACPrefix + strings.TrimPrefix(id, HACPrefix)
	}
	return id
}

func gameForRegion(id string, asset System, region Region, title, cover string) Game {
	g := Game{
		ID:          id,
		AssetSystem: asset,
		Icons:       map[Region]bool{},
		Titles:      map[Region]string{},
		CoverArt:    cover,
	}
	if title != "" {
		g.Titles[region] = title
		g.Icons[region] = true
	}
	return g
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

func (d euDoc) Game(want System, region Region) (Game, bool) {
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
	id = applyIDPrefix(id, want, asset)
	cover := d.ImageSq
	if cover == "" {
		cover = d.Image
	}
	return gameForRegion(id, asset, region, title, cover), true
}

func jpAsset(hard, sform string) System {
	u := strings.ToUpper(hard + " " + sform)
	switch {
	case strings.Contains(u, "BEE"):
		return BEE
	case strings.Contains(u, "HAC"):
		return HAC
	case strings.Contains(u, "CTR") || strings.Contains(u, "3DS"):
		return CTR
	case strings.Contains(u, "WUP") || strings.Contains(u, "WIIU"):
		return WUP
	default:
		return ""
	}
}

func (d jpItem) Game(want System) (Game, bool) {
	title := strings.TrimSpace(d.Title)
	if title == "" {
		return Game{}, false
	}
	asset := jpAsset(d.Hard, d.Sform)
	if !matchesSystem(want, asset) {
		return Game{}, false
	}
	id := strings.TrimSpace(d.Nsuid)
	if id == "" {
		id = strings.TrimSpace(d.ID)
	}
	if id == "" {
		return Game{}, false
	}
	id = applyIDPrefix(id, want, asset)
	cover := ""
	if d.IURL != "" {
		cover = jpImageBase + d.IURL + ".jpg"
	}
	return gameForRegion(id, asset, JP, title, cover), true
}

func usAsset(platform, code string) System {
	c := strings.ToUpper(strings.TrimSpace(code))
	if sys, ok := ParseSystem(c); ok {
		return sys
	}
	p := strings.ToLower(platform)
	switch {
	case strings.Contains(p, "switch 2"):
		return BEE
	case strings.Contains(p, "switch"):
		return HAC
	case strings.Contains(p, "3ds"):
		return CTR
	case strings.Contains(p, "wii u"):
		return WUP
	default:
		return ""
	}
}

func (d usHit) Game(want System) (Game, bool) {
	title := strings.TrimSpace(d.Title)
	id := strings.TrimSpace(d.Nsuid)
	if title == "" || id == "" || d.DLCType != "" {
		return Game{}, false
	}
	asset := usAsset(d.Platform, d.PlatformCode)
	if !matchesSystem(want, asset) {
		return Game{}, false
	}
	id = applyIDPrefix(id, want, asset)
	cover := d.ProductImageSquare
	if cover == "" && d.ProductImage != "" {
		if strings.HasPrefix(d.ProductImage, "http") {
			cover = d.ProductImage
		} else {
			cover = usImageBase + strings.TrimPrefix(d.ProductImage, "/")
		}
	}
	return gameForRegion(id, asset, US, title, cover), true
}

// SearchStore looks up titles on the eShop for the given preferred region.
func (c *Client) SearchStore(ctx context.Context, query string, system System, region Region) ([]Game, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return nil, nil
	}
	if c.StoreURL != "" {
		return c.searchEurope(ctx, query, system, region)
	}
	switch region {
	case JP:
		return c.searchJapan(ctx, query, system)
	case US:
		return c.searchAmerica(ctx, query, system)
	default:
		return c.searchEurope(ctx, query, system, EU)
	}
}

func (c *Client) searchEurope(ctx context.Context, query string, system System, region Region) ([]Game, error) {
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
	if region == "" {
		region = EU
	}
	out := make([]Game, 0, len(resp.Response.Docs))
	seen := map[string]bool{}
	for _, doc := range resp.Response.Docs {
		g, ok := doc.Game(system, region)
		if !ok || seen[g.ID] {
			continue
		}
		seen[g.ID] = true
		out = append(out, g)
	}
	return out, nil
}

func (c *Client) searchJapan(ctx context.Context, query string, system System) ([]Game, error) {
	u, err := url.Parse(defaultJPStore)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("opt_sshow", "1")
	q.Set("limit", "12")
	q.Set("page", "1")
	q.Set("q", query)
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
	var resp jpResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	out := make([]Game, 0, len(resp.Result.Items))
	seen := map[string]bool{}
	for _, item := range resp.Result.Items {
		g, ok := item.Game(system)
		if !ok || seen[g.ID] {
			continue
		}
		seen[g.ID] = true
		out = append(out, g)
	}
	return out, nil
}

func (c *Client) searchAmerica(ctx context.Context, query string, system System) ([]Game, error) {
	payload, err := json.Marshal(usQuery{Query: query, HitsPerPage: 12, Page: 0})
	if err != nil {
		return nil, err
	}
	body, err := c.do(ctx, http.MethodPost, usAlgoliaURL, payload, map[string]string{
		"Content-Type":             "application/json",
		"X-Algolia-API-Key":        usAlgoliaKey,
		"X-Algolia-Application-Id": usAlgoliaApp,
	})
	if err != nil {
		return nil, fmt.Errorf("eshop search: %w", err)
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var resp usResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	out := make([]Game, 0, len(resp.Hits))
	seen := map[string]bool{}
	for _, hit := range resp.Hits {
		g, ok := hit.Game(system)
		if !ok || seen[g.ID] {
			continue
		}
		seen[g.ID] = true
		out = append(out, g)
	}
	return out, nil
}

// EnrichRegions fills missing regional titles by searching the other stores.
func (c *Client) EnrichRegions(ctx context.Context, g Game, system System) (Game, error) {
	if g.Titles == nil {
		g.Titles = map[Region]string{}
	}
	if g.Icons == nil {
		g.Icons = map[Region]bool{}
	}
	seed := g.Title("")
	if seed == "" {
		return g, nil
	}
	for _, region := range Regions {
		if g.Titles[region] != "" {
			continue
		}
		hits, err := c.SearchStore(ctx, seed, system, region)
		if err != nil || len(hits) == 0 {
			continue
		}
		for _, h := range hits {
			ht := h.Title(region)
			if ht == "" {
				ht = h.Title("")
			}
			if !titlesSimilar(seed, ht) {
				continue
			}
			if t := h.Titles[region]; t != "" {
				g.Titles[region] = t
				g.Icons[region] = true
			}
			if g.CoverArt == "" {
				g.CoverArt = h.CoverArt
			}
			break
		}
	}
	return g, nil
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
