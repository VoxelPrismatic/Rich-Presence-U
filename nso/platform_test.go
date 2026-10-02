package nso

import (
	"testing"

	"gorm.io/gorm"
)

func TestFriendCodeList(t *testing.T) {
	var list FriendCodeList
	if err := list.Scan(nil); err != nil || list != nil {
		t.Fatalf("nil scan: %v %v", err, list)
	}
	if err := list.Scan([]byte("aa\x98bb\x98")); err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[0] != "aa" || list[1] != "bb" || list[2] != "" {
		t.Fatalf("bytes scan: %#v", list)
	}
	v, err := list.Value()
	if err != nil || v != "aa\x98bb\x98" {
		t.Fatalf("value %v %v", v, err)
	}
	empty := FriendCodeList(nil)
	v, err = empty.Value()
	if err != nil || v != nil {
		t.Fatalf("empty value %v %v", v, err)
	}
}

func TestHistoryTimeRank(t *testing.T) {
	if _, ok := HistoryRank(HistoryTime(-1)); ok {
		t.Fatal("floor time should stay out of the recent list")
	}
	for rank := 0; rank < 8; rank++ {
		got, ok := HistoryRank(HistoryTime(rank))
		if !ok || got != rank {
			t.Fatalf("rank %d -> %d %v", rank, got, ok)
		}
	}
}

func TestSaveLoadPlatforms(t *testing.T) {
	c, err := New(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })

	prefs := []PlatformPrefs{{
		Platform:     "NS1",
		GameID:       "newer",
		FriendCode:   FriendCodeList{"1111", "2222", "3333", ""},
		FcAsIcon:     true,
		PreserveTime: true,
	}}
	history := []PlatformHistory{
		{
			Model:        gorm.Model{CreatedAt: HistoryTime(-1), UpdatedAt: HistoryTime(-1)},
			Platform:     "NS1",
			GameID:       "old",
			Mode:         1,
			Description:  "kept",
			PartyEnabled: true,
			PartySize:    2,
			PartyMax:     4,
			Region:       RegionCode(EU),
		},
		{
			Model:    gorm.Model{CreatedAt: HistoryTime(0), UpdatedAt: HistoryTime(0)},
			Platform: "NS1",
			GameID:   "older-recent",
			Mode:     0,
		},
		{
			Model:       gorm.Model{CreatedAt: HistoryTime(1), UpdatedAt: HistoryTime(1)},
			Platform:    "NS1",
			GameID:      "newer",
			Mode:        2,
			Description: "fc",
			Region:      RegionCode(US),
		},
	}
	if err := c.SavePlatforms(prefs, history); err != nil {
		t.Fatal(err)
	}
	gotPrefs, gotHist, err := c.LoadPlatforms()
	if err != nil {
		t.Fatal(err)
	}
	if len(gotPrefs) != 1 || gotPrefs[0].GameID != "newer" || !gotPrefs[0].FcAsIcon || !gotPrefs[0].PreserveTime {
		t.Fatalf("prefs %+v", gotPrefs)
	}
	if len(gotPrefs[0].FriendCode) != 4 || gotPrefs[0].FriendCode[0] != "1111" || gotPrefs[0].FriendCode[2] != "3333" {
		t.Fatalf("friend code %#v", gotPrefs[0].FriendCode)
	}
	if len(gotHist) != 3 {
		t.Fatalf("history %+v", gotHist)
	}
	if gotHist[0].GameID != "old" || gotHist[1].GameID != "older-recent" || gotHist[2].GameID != "newer" {
		t.Fatalf("order %s %s %s", gotHist[0].GameID, gotHist[1].GameID, gotHist[2].GameID)
	}
	if _, ok := HistoryRank(gotHist[0].UpdatedAt); ok {
		t.Fatalf("old row should not be recent: %v", gotHist[0].UpdatedAt)
	}
	if rank, ok := HistoryRank(gotHist[2].UpdatedAt); !ok || rank != 1 {
		t.Fatalf("newest rank %d %v at %v", rank, ok, gotHist[2].UpdatedAt)
	}
	if gotHist[0].Description != "kept" || !gotHist[0].PartyEnabled || gotHist[0].PartySize != 2 || gotHist[0].Region != RegionCode(EU) {
		t.Fatalf("old row %+v", gotHist[0])
	}
}
