package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/covrom/dingo/app/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func mockPost() *Post {
	p := NewPost()
	p.Title = "Welcome to Dingo!"
	p.Slug = "welcome-to-dingo"
	p.Markdown = samplePostContent
	p.Html = utils.Markdown2Html(p.Markdown)
	p.IsPage = false
	p.AllowComment = true
	p.Category = ""
	p.CreatedBy = 0
	p.UpdatedBy = 0
	p.IsPublished = true
	return p
}

func TestPost(t *testing.T) {
	Convey("Initialize database", t, func() {
		testDB := fmt.Sprintf(filepath.Join(os.TempDir(), "ding-testdb-%s"), fmt.Sprintf(time.Now().Format("20060102T150405.000")))
		Initialize(testDB, true)

		Convey("Create a published post", func() {
			p := mockPost()
			tags := GenerateTagsFromCommaString("Welcome, Dingo")
			err := p.Save(tags...)

			So(err, ShouldBeNil)

			So(p.Id, ShouldEqual, 1)

			So(p.TagString(), ShouldEqual, "Welcome, Dingo")

			So(p.CreatedAt, ShouldNotBeNil)

			So(p.UpdatedAt, ShouldNotBeNil)

			So(p.PublishedAt, ShouldNotBeNil)

			Convey("Update post tag", func() {
				updateTags := GenerateTagsFromCommaString("Welcome")
				err = p.Save(updateTags...)

				So(err, ShouldBeNil)

				Convey("Unused tag should be removed", func() {
					tag := &Tag{Slug: "dingo"}
					err = tag.GetTagBySlug()
					So(err, ShouldNotBeNil)
				})

				Convey("Tags should be updated", func() {
					newPost := new(Post)
					newPost.Id = 1
					err := newPost.GetPostById()

					So(err, ShouldBeNil)
					tags := new(Tags)
					err = tags.GetTagsByPostId(p.Id)
					So(tags, ShouldHaveLength, 1)
					So(tags.Get(0).Slug, ShouldEqual, "welcome")
					//					So((*newPost.UpdatedAt).After(*p.UpdatedAt), ShouldBeTrue)
				})
			})

			Convey("Update post slug", func() {
				newSlug := "slug-modified"
				p.Slug = newSlug
				err = p.Save()

				So(err, ShouldBeNil)

				Convey("Slug should be updated", func() {
					p := &Post{Id: 1}
					err := p.GetPostById()

					So(err, ShouldBeNil)
					So(p.Slug, ShouldEqual, newSlug)
				})
			})

			Convey("Update post title", func() {
				newTitle := "Title modified"
				p.Title = newTitle
				err = p.Save()

				So(err, ShouldBeNil)

				Convey("Title should be updated", func() {
					newPost := new(Post)
					newPost.Id = 1
					err := newPost.GetPostById()

					So(err, ShouldBeNil)
					So(newPost.Title, ShouldEqual, newTitle)
				})
			})

			Convey("Delete post by ID", func() {
				DeletePostById(1)
				p := &Post{Id: 1}
				err := p.GetPostById()

				So(err, ShouldNotBeNil)

				Convey("Tags should be deleted", func() {
					tags := new(Tags)
					_ = tags.GetAllTags()

					So(tags, ShouldHaveLength, 0)
				})
			})

			Convey("Get post by Tag", func() {
				posts := new(Posts)
				pager, err := posts.GetPostsByTag(1, 1, 1, false)

				So(posts, ShouldHaveLength, 1)
				So(pager.Begin, ShouldEqual, 0)
				So(err, ShouldBeNil)
			})

			Convey("Get all posts by Tag", func() {
				posts := new(Posts)
				err := posts.GetAllPostsByTag(1)

				So(posts, ShouldHaveLength, 1)
				So(err, ShouldBeNil)
			})

			Convey("Get number of Posts", func() {
				num, err := GetNumberOfPosts(false, true)

				So(num, ShouldEqual, 1)
				So(err, ShouldBeNil)
			})

			Convey("Get post list", func() {
				posts := new(Posts)
				pager, err := posts.GetPostList(1, 1, false, false, "created_at")

				So(posts, ShouldHaveLength, 1)
				So(pager.Begin, ShouldEqual, 0)
				So(err, ShouldBeNil)
			})

			Convey("Create a post with the same slug", func() {
				newPost := mockPost()
				err := newPost.Save()

				So(err, ShouldBeNil)
				So(newPost.Slug, ShouldEqual, "welcome-to-dingo-1")

				Convey("Create a post with the same slug", func() {
					newPost := mockPost()
					err := newPost.Save()

					So(err, ShouldBeNil)
					So(newPost.Slug, ShouldEqual, "welcome-to-dingo-2")
				})
			})

		})

		Convey("Create welcome data", func() {
			createWelcomeData()

			Convey("Get the welcome post", func() {
				post := new(Post)
				post.Id = 1
				err := post.GetPostById()

				So(err, ShouldBeNil)
				So(post.Title, ShouldEqual, "Welcome to Dingo!")
			})
		})

		Reset(func() {
			os.Remove(testDB)
		})
	})
}
