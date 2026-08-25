package igdb

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	advancedSearchBase = "https://www.igdb.com/advanced_search"
	gamePageBase       = "https://www.igdb.com/games/"
	coverImageBase     = "https://images.igdb.com/igdb/image/upload/t_cover_big/"
)

// GameSearchURL is the public IGDB advanced-search page for a title on a
// console. Only the query value is escaped; filter names like f[type] and
// f[platforms.id_in] keep their brackets.
func GameSearchURL(query string, platformID PlatformID) string {
	var b strings.Builder
	b.Grow(len(advancedSearchBase) + 64 + len(query))
	b.WriteString(advancedSearchBase)
	b.WriteString("?d=1")
	b.WriteString("&f[type]=games")
	b.WriteString("&q=")
	b.WriteString(url.QueryEscape(strings.TrimSpace(query)))
	b.WriteString("&s=score")
	if platformID != 0 {
		b.WriteString("&f[platforms.id_in]=")
		b.WriteString(strconv.Itoa(int(platformID)))
	}
	return b.String()
}

// GamePageURL is the IGDB website page for a game slug.
func GamePageURL(slug string) string {
	slug = strings.Trim(strings.TrimSpace(slug), "/")
	if slug == "" {
		return ""
	}
	return gamePageBase + url.PathEscape(slug)
}

// CoverImageURL is a full-size cover from an IGDB image id.
func CoverImageURL(imageID string) string {
	imageID = strings.TrimSpace(imageID)
	if imageID == "" {
		return ""
	}
	return coverImageBase + imageID + ".jpg"
}

// FilterIDs is this console plus the IDs in Includes (eg NES + Famicom).
func (p Platform) FilterIDs() []int {
	ids := make([]int, 0, 1+len(p.Includes))
	seen := map[int]bool{}
	add := func(id int) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	add(int(p.ID))
	for _, id := range p.Includes {
		add(id)
	}
	return ids
}
