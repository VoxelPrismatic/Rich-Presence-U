package discord

import "strings"

// Activity is a Discord SET_ACTIVITY payload.
type Activity struct {
	Details        string
	State          string
	StartTimestamp int64
	EndTimestamp   int64
	LargeImage     string
	LargeText      string
	SmallImage     string
	SmallText      string
	PartySize      int
	PartyMax       int
}

// Presence is the app-level view of what should show on Discord.
type Presence struct {
	Title       string
	Description string
	Tag         string
	ShowTag     bool
	TagIcon     bool
	CoverKey    string
	Party       bool
	PartySize   int
	PartyMax    int
	Start       int64
	End         int64
}

func pad2(s string) string {
	s = strings.TrimRight(s, " ")
	if s == "" {
		return ""
	}
	if len([]rune(s)) < 2 {
		return s + "  "
	}
	return s
}

func blank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Build maps the editor fields onto a Discord activity the same way the
// original Godot app did (without the dropped "minimal status" layout).
func Build(p Presence) Activity {
	title := pad2(p.Title)
	desc := ""
	if !blank(p.Description) {
		desc = pad2(p.Description)
	}

	a := Activity{
		Details:        title,
		State:          desc,
		StartTimestamp: p.Start,
		EndTimestamp:   p.End,
		LargeImage:     p.CoverKey,
		LargeText:      "Rich Presence Qt",
	}
	if a.LargeImage == "" {
		a.LargeImage = "default"
	}

	if p.ShowTag && p.Tag != "" {
		if p.TagIcon || desc != "" {
			a.SmallText = p.Tag
			a.SmallImage = "id"
		} else {
			a.State = p.Tag
		}
	}

	if p.Party {
		if a.State == "" {
			a.State = "  "
		}
		a.PartySize = p.PartySize
		a.PartyMax = p.PartyMax
		if a.PartyMax < 1 {
			a.PartyMax = 1
		}
		if a.PartySize < 1 {
			a.PartySize = 1
		}
		if a.PartySize > a.PartyMax {
			a.PartySize = a.PartyMax
		}
	}
	return a
}

func (a Activity) payload() map[string]any {
	out := map[string]any{"instance": true}
	if a.Details != "" {
		out["details"] = a.Details
	}
	if a.State != "" {
		out["state"] = a.State
	}
	ts := map[string]any{}
	if a.StartTimestamp > 0 {
		ts["start"] = a.StartTimestamp
	}
	if a.EndTimestamp > 0 {
		ts["end"] = a.EndTimestamp
	}
	if len(ts) > 0 {
		out["timestamps"] = ts
	}
	assets := map[string]any{}
	if a.LargeImage != "" {
		assets["large_image"] = a.LargeImage
	}
	if a.LargeText != "" {
		assets["large_text"] = a.LargeText
	}
	if a.SmallImage != "" {
		assets["small_image"] = a.SmallImage
	}
	if a.SmallText != "" {
		assets["small_text"] = a.SmallText
	}
	if len(assets) > 0 {
		out["assets"] = assets
	}
	if a.PartyMax > 0 {
		out["party"] = map[string]any{
			"size": []int{a.PartySize, a.PartyMax},
		}
	}
	return out
}
