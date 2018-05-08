package model

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func mockTag(name, slug string) Tag {
	return NewTag(name, slug)
}

func tagEqualCheck(tag, expected Tag) {
	So(tag.Name, ShouldEqual, expected.Name)
	So(tag.Slug, ShouldEqual, expected.Slug)
}

func TestTag(t *testing.T) {
	Convey("Initialize database", t, func() {
		DBName = fmt.Sprintf("ding-testdb-%s", time.Now().Format("20060102T150405"))
		Initialize("localhost", true)

		Convey("Test Tag", func() {
			p := mockPost()
			tags := GenerateTagsFromCommaString("Welcome, Dingo")

			name := "test-tag"
			slug := GenerateSlug(name, "tags")

			tag := mockTag(name, slug)

			tags = append(tags, tag)

			err := p.Save(tags...)
			// err = tag.Save()
			// So(err, ShouldBeNil)

			Convey("Geenrate tags from comma string", func() {
				tags := GenerateTagsFromCommaString("tagA,tagB")

				So(tags, ShouldHaveLength, 2)
			})

			Convey("Get tags by post ID", func() {
				tags := new(Tags)
				err = tags.GetTagsByPostId(p.Id.Hex())

				So(tags, ShouldHaveLength, 3)
				So(err, ShouldBeNil)
			})

			// Convey("Get tag", func() {
			// 	t := &Tag{Id: tag.Id}
			// 	err := t.GetTag()

			// 	tagEqualCheck(t, tag)
			// 	So(err, ShouldBeNil)
			// })

			Convey("Get tag by slug", func() {
				ttag := &Tag{}
				*ttag = p.Tags[0]
				err = ttag.GetTagBySlug()

				tagEqualCheck(*ttag, p.Tags[0])
				So(err, ShouldBeNil)
			})

			Convey("Get all tags", func() {
				ts := new(Tags)
				err := ts.GetAllTags()

				// fmt.Printf("tags: %#v\n", ts)

				So(ts, ShouldHaveLength, 3)
				So(err, ShouldBeNil)
			})

			// Convey("Delete old tags", func() {
			// 	err := DeleteOldTags()
			// 	So(err, ShouldBeNil)

			// 	ts := new(Tags)
			// 	err = ts.GetAllTags()
			// 	So(err, ShouldBeNil)
			// 	// delete test-tag created by mockTag(),
			// 	// but two tags created by mockPost remain.
			// 	So(ts, ShouldHaveLength, 2)
			// })

		})
		Reset(func() {
			DropDatabase()
		})
	})
}
