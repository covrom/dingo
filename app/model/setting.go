package model

import (
	"encoding/json"
	"time"

	"github.com/covrom/dingo/app/utils"
	"github.com/globalsign/mgo/bson"
)

// const stmtGetSetting = `SELECT * FROM settings WHERE key = ?`
// const stmtSaveSelect = `SELECT id FROM settings WHERE KEY = ?`
// const stmtGetSettingsByType = `SELECT * FROM settings WHERE type = ?`

// A Setting is the data type that stores the blog's configuration options. It
// is essentially a key-value store for settings, along with a type to help
// specify the specific type of setting. A type can be either
//         general        site-wide general settings
//         content        related to showing content
//         navigation     site navigation settings
//         custom         custom settings
type Setting struct {
	// Id        int        `meddler:"id,pk"`
	Key       string
	Value     string
	Type      string // general, content, navigation, custom
	CreatedAt *time.Time
	CreatedBy int64
	UpdatedAt *time.Time
	UpdatedBy int64
}

// A Navigator represents a link in the site navigation menu.
type Navigator struct {
	Label string `json:"label"`
	Url   string `json:"url"`
}

// GetNavigators returns a slice of all Navigators.
func GetNavigators() []*Navigator {
	var navs []*Navigator
	navStr := GetSettingValue("navigation")
	json.Unmarshal([]byte(navStr), &navs)
	return navs
}

// SetNavigators saves one or more label-url pairs in the site's Settings.
func SetNavigators(labels, urls []string) error {
	var navs []*Navigator
	for i, l := range labels {
		if len(l) < 1 {
			continue
		}
		navs = append(navs, &Navigator{l, urls[i]})
	}
	navStr, err := json.Marshal(navs)
	if err != nil {
		return err
	}

	s := NewSetting("navigation", string(navStr), "navigation")
	return s.Save()
}

// GetSetting checks if a setting exists in the DB.
func (setting *Setting) GetSetting() error {
	// session := mdb.Copy()
	// defer session.Close()

	err := setSession.Clone().DB(DBName).C("settings").Find(bson.M{"key": setting.Key}).One(setting)

	return err
}

// GetSettingValue returns the Setting value associated with the given Setting
// key.
func GetSettingValue(k string) string {
	// TODO: error handling
	setting := &Setting{Key: k}
	_ = setting.GetSetting()
	return setting.Value
}

// GetCustomSettings returns all custom settings.
func GetCustomSettings() *Settings {
	return GetSettingsByType("custom")
}

// Settings a slice of all "Setting"s
type Settings []*Setting

// GetSettingsByType returns all settings of the given type, where the setting
// key can be one of "general", "content", "navigation", or "custom".
func GetSettingsByType(t string) *Settings {
	// session := mdb.Copy()
	// defer session.Close()

	settings := new(Settings)
	err := setSession.Clone().DB(DBName).C("settings").Find(bson.M{"type": t}).All(settings)

	if err != nil {
		return nil
	}
	return settings
}

// Save saves the setting to the DB.
func (setting *Setting) Save() error {
	// session := mdb.Copy()
	// defer session.Close()
	_, err := setSession.Clone().DB(DBName).C("settings").Upsert(bson.M{"key": setting.Key}, setting)
	return err
}

// NewSetting returns a new setting from the given key-value pair.
func NewSetting(k, v, t string) *Setting {
	return &Setting{
		Key:       k,
		Value:     v,
		Type:      t,
		CreatedAt: utils.Now(),
	}
}

// SetSettingIfNotExists sets the setting created by the given key-value pair
// if the setting does not yet exist.
func SetSettingIfNotExists(k, v, t string) error {
	s := NewSetting(k, v, t)
	err := s.GetSetting()
	if err != nil {
		s := NewSetting(k, v, t)
		return s.Save()
	}
	return err
}
