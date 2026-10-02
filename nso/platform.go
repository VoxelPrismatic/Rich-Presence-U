package nso

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// platformHistoryFloor is older than any remembered game.
// Rows at this time are library state that is not in the recent list.
var platformHistoryFloor = time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

// platformHistoryRecent is the start of remembered-game timestamps.
// Later seconds are more recent.
var platformHistoryRecent = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// FriendCodeList is a list of friend-code parts stored in one text column.
// Parts are joined with 0x98, the same separator used for string lists in GORM.
type FriendCodeList []string

func (f *FriendCodeList) Scan(value any) error {
	if value == nil {
		*f = nil
		return nil
	}
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("friend code: unable to convert %T", value)
	}
	if str == "" {
		*f = nil
		return nil
	}
	*f = strings.Split(str, "\x98")
	return nil
}

func (f FriendCodeList) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	return strings.Join(f, "\x98"), nil
}

// PlatformHistory is one game's saved status on a platform.
// UpdatedAt orders the recent list, newest last.
type PlatformHistory struct {
	gorm.Model
	Platform     string `gorm:"uniqueIndex:idx_platform_game,priority:1;size:64"`
	GameID       string `gorm:"uniqueIndex:idx_platform_game,priority:2;size:128"`
	Mode         int
	Description  string
	PartyEnabled bool
	PartySize    int
	PartyMax     int
	Region       int
}

// PlatformPrefs is the selected game and friend-code settings for one platform.
type PlatformPrefs struct {
	gorm.Model
	Platform     string         `gorm:"uniqueIndex;size:64"`
	GameID       string         `gorm:"size:128"`
	FriendCode   FriendCodeList `gorm:"type:text"`
	FcAsIcon     bool
	PreserveTime bool
}

// RegionCode stores a storefront region as an int. 0 is unset.
func RegionCode(r Region) int {
	switch r {
	case US:
		return 1
	case EU:
		return 2
	case JP:
		return 3
	default:
		return 0
	}
}

// RegionFromCode reverses RegionCode.
func RegionFromCode(code int) Region {
	switch code {
	case 1:
		return US
	case 2:
		return EU
	case 3:
		return JP
	default:
		return ""
	}
}

// HistoryTime is the UpdatedAt for a remembered game. rank 0 is the oldest.
func HistoryTime(rank int) time.Time {
	if rank < 0 {
		return platformHistoryFloor
	}
	return platformHistoryRecent.Add(time.Duration(rank) * time.Second)
}

// HistoryRank reports whether updatedAt belongs to the recent list.
// The second result is the oldest-first position.
func HistoryRank(updatedAt time.Time) (int, bool) {
	if updatedAt.IsZero() {
		return 0, false
	}
	recent := time.Date(2000, 1, 1, 0, 0, 0, 0, updatedAt.Location())
	if updatedAt.Before(recent) {
		return 0, false
	}
	return int(updatedAt.Sub(recent) / time.Second), true
}

// LoadPlatforms reads platform prefs and game history, oldest updated_at first.
func (c *Client) LoadPlatforms() ([]PlatformPrefs, []PlatformHistory, error) {
	if c == nil || c.db == nil {
		return nil, nil, fmt.Errorf("games.db is closed")
	}
	var prefs []PlatformPrefs
	if err := c.db.Order("updated_at asc").Find(&prefs).Error; err != nil {
		return nil, nil, err
	}
	var hist []PlatformHistory
	if err := c.db.Order("updated_at asc, game_id asc").Find(&hist).Error; err != nil {
		return nil, nil, err
	}
	return prefs, hist, nil
}

// SavePlatforms replaces the platform tables. History UpdatedAt values are kept
// so the recent list still sorts after a reload.
func (c *Client) SavePlatforms(prefs []PlatformPrefs, history []PlatformHistory) error {
	if c == nil || c.db == nil {
		return fmt.Errorf("games.db is closed")
	}
	return c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1 = 1").Delete(&PlatformPrefs{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("1 = 1").Delete(&PlatformHistory{}).Error; err != nil {
			return err
		}
		if len(prefs) > 0 {
			if err := tx.Create(&prefs).Error; err != nil {
				return err
			}
		}
		if len(history) == 0 {
			return nil
		}
		stamps := make([]time.Time, len(history))
		for i := range history {
			stamps[i] = history[i].UpdatedAt
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		for i := range history {
			if stamps[i].IsZero() {
				continue
			}
			err := tx.Model(&PlatformHistory{}).
				Where("platform = ? AND game_id = ?", history[i].Platform, history[i].GameID).
				UpdateColumn("updated_at", stamps[i]).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
