package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/covrom/dingo/app/utils"
	"github.com/globalsign/mgo/bson"
	. "github.com/smartystreets/goconvey/convey"
)

func mockComment(id1, id2 bson.ObjectId) *Comment {
	c := NewComment()
	c.Author = name
	c.Email = email
	c.Website = "http://example.com"
	c.Content = "comment test"
	c.Avatar = utils.Gravatar(c.Email, "50")
	c.Parent = ""
	c.PostId = string(id2)
	//	c.Ip = "127.0.0.1"
	c.UserAgent = "Mozilla"
	c.UserId = string(id1)
	c.Approved = true
	return c
}

func commentEqualCheck(c *Comment, expected *Comment) {
	So(c.Author, ShouldEqual, expected.Author)
	So(c.Email, ShouldEqual, expected.Email)
	So(c.Website, ShouldEqual, expected.Website)
	So(c.Content, ShouldEqual, expected.Content)
	So(c.Avatar, ShouldEqual, expected.Avatar)
	So(c.Parent, ShouldEqual, expected.Parent)
	So(c.PostId, ShouldEqual, expected.PostId)
	So(c.Ip, ShouldEqual, expected.Ip)
	So(c.UserAgent, ShouldEqual, expected.UserAgent)
	So(c.UserId, ShouldEqual, expected.UserId)
	So(c.Approved, ShouldEqual, expected.Approved)
}

func TestComment(t *testing.T) {
	id1 := bson.NewObjectId()
	id2 := bson.NewObjectId()
	Convey("Initialize database", t, func() {
		DBName = fmt.Sprintf("ding-testdb-%s", time.Now().Format("20060102T150405"))
		Initialize("localhost")

		Convey("Test Message", func() {
			pc := mockComment(id1, id2)
			err := pc.Save()
			So(err, ShouldBeNil)

			cc := mockComment(id1, id2)
			cc.Parent = string(pc.Id)
			cc.Content = "comment test by child"
			err = cc.Save()
			So(err, ShouldBeNil)

			Convey("Get Comment List", func() {
				comments := new(Comments)
				_, err := comments.GetCommentList(1, 2, false)
				So(err, ShouldBeNil)
				So(comments, ShouldHaveLength, 2)
			})
			Convey("To Json", func() {
				result := cc.ToJson()
				So(result, ShouldNotBeNil)
			})

			Convey("Parent Content", func() {
				result := cc.ParentContent()
				So(result, ShouldEqual, fmt.Sprintf("> @%s\n\n> %s\n", pc.Author, pc.Content))
			})

			Convey("Get Number of Comments", func() {
				result, err := GetNumberOfComments()
				So(err, ShouldBeNil)
				So(result, ShouldEqual, 2)
			})

			Convey("Get Comment By ID", func() {
				result := &Comment{Id: cc.Id}
				err := result.GetCommentById()
				So(err, ShouldBeNil)
				commentEqualCheck(result, cc)
			})

			Convey("Get Comments By Post ID", func() {
				comments := new(Comments)
				err := comments.GetCommentsByPostId(string(cc.Id))
				So(err, ShouldBeNil)
				commentEqualCheck(comments.Get(0), pc)
				commentEqualCheck(comments.Get(0).Children.Get(0), cc)

			})

			Convey("Validate Comment", func() {
				result := cc.ValidateComment()
				So(result, ShouldEqual, "")
			})

			Convey("Delete Comment", func() {
				err := DeleteComment(string(cc.Id))
				So(err, ShouldBeNil)
				result := &Comment{Id: cc.Id}
				err = result.GetCommentById()
				So(err, ShouldNotBeNil)
				So(result.CreatedAt, ShouldBeNil)
			})
		})
		Reset(func() {
			DropDatabase()
		})
	})
}
