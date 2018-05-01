package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetting(t *testing.T) {

	Convey("Initialize database", t, func() {
		DBName = fmt.Sprintf(filepath.Join(os.TempDir(), "ding-testdb-%s"), fmt.Sprintf(time.Now().Format("20060102T150405.000")))
		Initialize("localhost")

		Convey("Test Navigators", func() {
			SetNavigators([]string{"Home", "About", "Team", ""}, []string{"/", "/about/", "/team/", "/other/"})

			Convey("Get Navigators", func() {
				navs := GetNavigators()

				So(navs, ShouldHaveLength, 3)
				So(navs[0].Label, ShouldEqual, "Home")
				So(navs[0].Url, ShouldEqual, "/")
			})
		})

		Convey("Test Settings", func() {
			err := NewSetting("github", "dinever", "custom").Save()
			So(err, ShouldBeNil)

			err = NewSetting("twitter", "Dingpeixuan", "custom").Save()
			So(err, ShouldBeNil)

			Convey("Get Setting", func() {
				s := &Setting{Key: "github"}
				err = s.GetSetting()

				So(err, ShouldBeNil)
				So(s.Key, ShouldEqual, "github")
				So(s.Value, ShouldEqual, "dinever")
				So(s.Type, ShouldEqual, "custom")
				So(s.CreatedAt, ShouldNotBeNil)
			})

			Convey("Set Setting if not exists", func() {
				SetSettingIfNotExists("github", "github", "custom")
				s := &Setting{Key: "github"}
				err = s.GetSetting()

				So(err, ShouldBeNil)
				So(s.Key, ShouldEqual, "github")
				So(s.Value, ShouldEqual, "dinever")
				So(s.Type, ShouldEqual, "custom")
			})

			Convey("Get custom settings", func() {
				cs := GetCustomSettings()

				So(cs, ShouldHaveLength, 2)
			})

			Convey("Get all settings", func() {
				cs := GetSettingsByType("custom")

				So(cs, ShouldHaveLength, 2)
			})

			Convey("Get setting value", func() {
				v := GetSettingValue("github")
				So(v, ShouldEqual, "dinever")

				v = GetSettingValue("gplus")
				So(v, ShouldEqual, "")
			})
		})

		Reset(func() {
			DropDatabase()
		})
	})
}
