package igdb

import (
	"sort"
	"strings"
	"unicode"
)

const (
	fzfMatch       = 10.0
	fzfBonus       = 10.0
	fzfConsecutive = 5.0
	fzfGapPenalty  = 3.0
	fzfGapLength   = 1.0
	fzfExactBonus  = 50.0
	chevron        = "  ›  "
)

// Hit is one platform picker search result, labeled as "Parent > Child".
type Hit struct {
	Platform Platform
	Parent   string
	Child    string
	Score    float64
}

// Label is the contextual row text, e.g. "Nintendo > GameCube".
func (h Hit) Label() string {
	if h.Parent == "" {
		return h.Child
	}
	return h.Parent + chevron + h.Child
}

// Search finds mapped consoles. The query is lowercased and split on spaces;
// every token must fuzzy-match the platform (names, slug, suffix, family).
func Search(query string) []Hit {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return nil
	}

	var hits []Hit
	for maker, families := range MappedPlatforms {
		for fam, plats := range families {
			for _, p := range plats {
				parent := fam
				if len(plats) == 1 {
					parent = maker
				}
				hay := searchHaystack(maker, fam, p)
				score, ok := matchTerms(hay, terms)
				if !ok {
					continue
				}
				hits = append(hits, Hit{
					Platform: p,
					Parent:   parent,
					Child:    p.DisplayName(),
					Score:    score,
				})
			}
		}
	}

	hits = dedupeHits(hits)
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].Label() < hits[j].Label()
	})
	return hits
}

// dedupeHits keeps one result per platform ID (or slug if ID is unset).
// The higher-scoring placement wins so a query like "virtual reality"
// prefers the VR tree copy of PSVR over the PlayStation copy.
func dedupeHits(hits []Hit) []Hit {
	type key struct {
		id   PlatformID
		slug string
	}
	best := make(map[key]Hit, len(hits))
	order := make([]key, 0, len(hits))
	for _, h := range hits {
		k := key{id: h.Platform.ID}
		if k.id == 0 {
			k.slug = h.Platform.Slug
		}
		prev, ok := best[k]
		if !ok {
			best[k] = h
			order = append(order, k)
			continue
		}
		if h.Score > prev.Score || (h.Score == prev.Score && h.Label() < prev.Label()) {
			best[k] = h
		}
	}
	out := make([]Hit, 0, len(order))
	for _, k := range order {
		out = append(out, best[k])
	}
	return out
}

func searchHaystack(maker, fam string, p Platform) string {
	parts := []string{maker, fam, p.Slug, p.DisplayName()}
	parts = append(parts, p.Names...)
	parts = append(parts, p.Suffix...)
	for _, n := range p.Names {
		for _, s := range p.Suffix {
			parts = append(parts, n+" "+s)
		}
	}
	for i, s := range parts {
		parts[i] = strings.ToLower(s)
	}
	return strings.Join(parts, " ")
}

func matchTerms(hay string, terms []string) (float64, bool) {
	total := 0.0
	for _, term := range terms {
		score, ok := fuzzyScore(term, hay)
		if !ok {
			return 0, false
		}
		total += score
	}
	return total, true
}

func fuzzyScore(term, text string) (float64, bool) {
	if term == "" {
		return 0, true
	}
	tr := []rune(term)
	tx := []rune(text)
	if pos := runeIndex(tx, tr); pos >= 0 {
		return scoreRunes(tx, rangePositions(pos, len(tr))) + fzfExactBonus, true
	}
	pos := shortestSubseq(tr, tx)
	if pos == nil {
		return 0, false
	}
	return scoreRunes(tx, pos), true
}

func runeIndex(text, pat []rune) int {
	if len(pat) == 0 || len(pat) > len(text) {
		return -1
	}
outer:
	for i := 0; i+len(pat) <= len(text); i++ {
		for j, r := range pat {
			if text[i+j] != r {
				continue outer
			}
		}
		return i
	}
	return -1
}

func rangePositions(start, n int) []int {
	pos := make([]int, n)
	for i := range pos {
		pos[i] = start + i
	}
	return pos
}

func shortestSubseq(term, text []rune) []int {
	var best []int
	bestSpan := 0
	for start := 0; start < len(text); start++ {
		if text[start] != term[0] {
			continue
		}
		attempt := make([]int, 0, len(term))
		attempt = append(attempt, start)
		ti := 1
		for i := start + 1; i < len(text) && ti < len(term); i++ {
			if text[i] == term[ti] {
				attempt = append(attempt, i)
				ti++
			}
		}
		if ti != len(term) {
			continue
		}
		span := attempt[len(attempt)-1] - attempt[0]
		if best == nil || span < bestSpan {
			best = attempt
			bestSpan = span
		}
	}
	return best
}

func scoreRunes(text []rune, pos []int) float64 {
	if len(pos) == 0 {
		return 0
	}
	score := 0.0
	for i, p := range pos {
		score += fzfMatch + fzfBonus*boundaryBonus(text, p)
		if i == 0 {
			continue
		}
		gap := p - pos[i-1]
		if gap == 1 {
			score += fzfConsecutive
		} else {
			score -= fzfGapPenalty + float64(gap)*fzfGapLength
		}
	}
	return score
}

func boundaryBonus(text []rune, i int) float64 {
	if i <= 0 {
		return 1
	}
	prev := text[i-1]
	if unicode.IsSpace(prev) || strings.ContainsRune("/_-.:", prev) {
		return 1
	}
	if unicode.IsLetter(prev) || unicode.IsDigit(prev) {
		return 0.5
	}
	return 0
}
