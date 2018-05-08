package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/covrom/dingo/app/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func mockPost() *Post {
	p := NewPost()
	// p.Id = tmp_post_id_1
	p.Title = "Welcome to Dingo!"
	p.Slug = "welcome-to-dingo"
	p.Markdown = "[test] "+samplePostContent
	p.Html = utils.Markdown2Html(p.Markdown)
	p.IsPage = false
	p.AllowComment = true
	p.Category = ""
	p.CreatedBy = ""
	p.UpdatedBy = ""
	p.IsPublished = true
	return p
}

func TestPost(t *testing.T) {
	Convey("Initialize database", t, func() {
		DBName = fmt.Sprintf("ding-testdb-%s", time.Now().Format("20060102T150405"))
		Initialize("localhost", true)

		Convey("Create a published post", func() {
			p := mockPost()
			tags := GenerateTagsFromCommaString("Welcome, Dingo")
			// t.Logf("%v\n", tags)
			err := p.Save(tags...)

			So(err, ShouldBeNil)

			// So(p.Id, ShouldEqual, 1)

			So(p.TagString(), ShouldEqual, "Welcome, Dingo")

			So(p.CreatedAt, ShouldNotBeNil)

			So(p.UpdatedAt, ShouldNotBeNil)

			So(p.PublishedAt, ShouldNotBeNil)

			updateTags := GenerateTagsFromCommaString("Welcome")

			Convey("Update post tag", func() {
				err = p.Save(updateTags...)

				So(err, ShouldBeNil)

				// Convey("Unused tag should be removed", func() {
				// 	tag := &Tag{Slug: "dingo"}
				// 	err = tag.GetTagBySlug()
				// 	// fmt.Printf("%#v\n", tag)
				// 	So(err, ShouldNotBeNil)
				// })

				Convey("Tags should be updated", func() {
					newPost := new(Post)
					newPost.Id = p.Id
					err := newPost.GetPostById()

					So(err, ShouldBeNil)

					tags := new(Tags)
					err = tags.GetTagsByPostId(p.Id.Hex())
					So(tags, ShouldHaveLength, 1)
					So(tags.Get(0).Slug, ShouldEqual, "welcome")
					//					So((*newPost.UpdatedAt).After(*p.UpdatedAt), ShouldBeTrue)
				})
			})

			Convey("Update post slug", func() {
				newSlug := "slug-modified"
				p.Slug = newSlug
				err = p.Save()

				// fmt.Printf("slug: %s", p.Slug)

				So(err, ShouldBeNil)

				Convey("Slug should be updated", func() {
					p2 := &Post{Id: p.Id}
					err := p2.GetPostById()

					So(err, ShouldBeNil)
					So(p2.Slug, ShouldEqual, newSlug)
				})
			})

			Convey("Update post title", func() {
				newTitle := "Title modified"
				p.Title = newTitle
				err = p.Save()

				So(err, ShouldBeNil)

				Convey("Title should be updated", func() {
					newPost := new(Post)
					newPost.Id = p.Id
					err := newPost.GetPostById()

					So(err, ShouldBeNil)
					So(newPost.Title, ShouldEqual, newTitle)
				})
			})

			Convey("Delete post by ID", func() {
				DeletePostById(tmp_post_id_1.Hex())
				p := &Post{Id: tmp_post_id_1}
				err := p.GetPostById()

				So(err, ShouldNotBeNil)

				// Convey("Tags should be deleted", func() {
				// 	tags := new(Tags)
				// 	_ = tags.GetAllTags()

				// 	So(tags, ShouldHaveLength, 0)
				// })
			})

			Convey("Get post by Tag", func() {
				posts := new(Posts)
				pager, err := posts.GetPostsByTag(updateTags[0], 1, 1, false)

				So(posts, ShouldHaveLength, 1)
				So(pager.Begin, ShouldEqual, 0)
				So(err, ShouldBeNil)
			})

			Convey("Get all posts by Tag", func() {
				posts := new(Posts)
				err := posts.GetAllPostsByTag(updateTags[0])

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
				post.Id = tmp_post_id_1
				err := post.GetPostById()

				So(err, ShouldBeNil)
				So(post.Title, ShouldEqual, "Welcome to Dingo!")
			})
		})

		Reset(func() {
			DropDatabase()
		})
	})
}
