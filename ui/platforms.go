package ui

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/voxelprismatic/richpresenceu/nso"
	"gorm.io/gorm"
)

const (
	modeEmpty  = 0
	modeCustom = 1
	modeFriend = 2
)

func encodeMode(mode string) int {
	switch mode {
	case "custom":
		return modeCustom
	case "friendcode":
		return modeFriend
	default:
		return modeEmpty
	}
}

func decodeMode(mode int) string {
	switch mode {
	case modeCustom:
		return "custom"
	case modeFriend:
		return "friendcode"
	default:
		return "empty"
	}
}

// encodeFriendCode stores the three friend-code boxes and the NNID.
// An empty code is a nil list. Otherwise the list is [part, part, part, nnid].
func encodeFriendCode(fc [3]string, tagID string) nso.FriendCodeList {
	if fc[0] == "" && fc[1] == "" && fc[2] == "" && tagID == "" {
		return nil
	}
	return nso.FriendCodeList{fc[0], fc[1], fc[2], tagID}
}

func decodeFriendCode(parts nso.FriendCodeList) (fc [3]string, tagID string) {
	if len(parts) > 0 {
		fc[0] = parts[0]
	}
	if len(parts) > 1 {
		fc[1] = parts[1]
	}
	if len(parts) > 2 {
		fc[2] = parts[2]
	}
	if len(parts) > 3 {
		tagID = parts[3]
	}
	return fc, tagID
}

func platformsToRows(systems map[string]*SystemState) ([]nso.PlatformPrefs, []nso.PlatformHistory) {
	keys := make([]string, 0, len(systems))
	for key := range systems {
		if key != "" && systems[key] != nil {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var prefs []nso.PlatformPrefs
	var hist []nso.PlatformHistory
	for _, key := range keys {
		st := systems[key]
		prefs = append(prefs, nso.PlatformPrefs{
			Platform:     key,
			GameID:       st.Game,
			FriendCode:   encodeFriendCode(st.TagFC, st.TagID),
			FcAsIcon:     st.TagIcon,
			PreserveTime: st.TimePreserve,
		})
		rank := map[string]int{}
		for i, id := range st.History {
			if id != "" {
				rank[id] = i
			}
		}
		ids := map[string]GameState{}
		for id, g := range st.Library {
			if id != "" {
				ids[id] = g
			}
		}
		for _, id := range st.History {
			if id == "" {
				continue
			}
			if _, ok := ids[id]; !ok {
				ids[id] = defaultGame()
			}
		}
		gameIDs := make([]string, 0, len(ids))
		for id := range ids {
			gameIDs = append(gameIDs, id)
		}
		sort.Strings(gameIDs)
		for _, id := range gameIDs {
			g := ids[id]
			ts := nso.HistoryTime(-1)
			if n, ok := rank[id]; ok {
				ts = nso.HistoryTime(n)
			}
			hist = append(hist, nso.PlatformHistory{
				Model:        gorm.Model{CreatedAt: ts, UpdatedAt: ts},
				Platform:     key,
				GameID:       id,
				Mode:         encodeMode(g.Mode),
				Description:  g.Description,
				PartyEnabled: g.Party,
				PartySize:    g.PartySize,
				PartyMax:     g.PartyMax,
				Region:       nso.RegionCode(g.Region),
			})
		}
	}
	return prefs, hist
}

func systemsFromDB(prefs []nso.PlatformPrefs, hist []nso.PlatformHistory) map[string]*SystemState {
	systems := map[string]*SystemState{}
	for _, p := range prefs {
		key := platformKey(p.Platform)
		if key == "" {
			continue
		}
		st := defaultSystem()
		st.Game = p.GameID
		st.TagFC, st.TagID = decodeFriendCode(p.FriendCode)
		st.TagIcon = p.FcAsIcon
		st.TimePreserve = p.PreserveTime
		systems[key] = &st
	}
	byPlatform := map[string][]nso.PlatformHistory{}
	for _, row := range hist {
		key := platformKey(row.Platform)
		if key == "" || row.GameID == "" {
			continue
		}
		byPlatform[key] = append(byPlatform[key], row)
	}
	for key, rows := range byPlatform {
		st := systems[key]
		if st == nil {
			d := defaultSystem()
			st = &d
			systems[key] = st
		}
		if st.Library == nil {
			st.Library = map[string]GameState{}
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].UpdatedAt.Equal(rows[j].UpdatedAt) {
				return rows[i].GameID < rows[j].GameID
			}
			return rows[i].UpdatedAt.Before(rows[j].UpdatedAt)
		})
		var recent []string
		for _, row := range rows {
			g := GameState{
				Mode:        decodeMode(row.Mode),
				Description: row.Description,
				Party:       row.PartyEnabled,
				PartySize:   row.PartySize,
				PartyMax:    row.PartyMax,
				Region:      nso.RegionFromCode(row.Region),
			}
			st.Library[row.GameID] = g
			if _, ok := nso.HistoryRank(row.UpdatedAt); ok {
				recent = append(recent, row.GameID)
			}
		}
		if len(recent) > 8 {
			recent = recent[len(recent)-8:]
		}
		st.History = recent
	}
	return systems
}

func prefsHavePlatforms(dir string) bool {
	b, err := os.ReadFile(prefsPath(dir))
	if err != nil {
		return false
	}
	var probe struct {
		Platforms json.RawMessage `json:"platforms"`
	}
	if json.Unmarshal(b, &probe) != nil {
		return false
	}
	raw := string(probe.Platforms)
	return raw != "" && raw != "null" && raw != "{}"
}

// platformsFieldPopulated reports that prefs.json still has platform rows to move.
func platformsFieldPopulated(systems map[string]*SystemState) bool {
	for key, st := range systems {
		if key != "" && st != nil {
			return true
		}
	}
	return false
}

// migratePlatforms copies a populated prefs.json platforms object into games.db
// and removes the field. A missing or empty field does nothing.
func migratePlatforms(client *nso.Client, settings Settings, fromFile map[string]*SystemState) (map[string]*SystemState, error) {
	if client == nil || !platformsFieldPopulated(fromFile) {
		return nil, nil
	}
	prefs, hist := platformsToRows(fromFile)
	if err := client.SavePlatforms(prefs, hist); err != nil {
		return nil, err
	}
	if err := savePrefs(client.ConfigDir, settings); err != nil {
		return nil, err
	}
	dbPrefs, dbHist, err := client.LoadPlatforms()
	if err != nil {
		return fromFile, nil
	}
	return systemsFromDB(dbPrefs, dbHist), nil
}

// loadPlatforms reads games.db. When prefs.json still has a platforms field,
// that field is migrated first.
func loadPlatforms(client *nso.Client, settings Settings, fromFile map[string]*SystemState) map[string]*SystemState {
	if fromFile == nil {
		fromFile = map[string]*SystemState{}
	}
	if client == nil {
		return fromFile
	}
	if platformsFieldPopulated(fromFile) {
		migrated, err := migratePlatforms(client, settings, fromFile)
		if err != nil || migrated == nil {
			return fromFile
		}
		return migrated
	}
	dbPrefs, dbHist, err := client.LoadPlatforms()
	if err != nil || (len(dbPrefs) == 0 && len(dbHist) == 0) {
		return fromFile
	}
	return systemsFromDB(dbPrefs, dbHist)
}
