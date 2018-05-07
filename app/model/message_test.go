package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	. "github.com/smartystreets/goconvey/convey"
)

func mockMessage(c *Comment) *Message {
	m := NewMessage("comment", c)
	return m
}

func TestMessage(t *testing.T) {
	// id1 := bson.NewObjectId()
	id2 := bson.NewObjectId()
	Convey("Initialize database", t, func() {
		DBName = fmt.Sprintf("ding-testdb-%s", time.Now().Format("20060102T150405"))
		Initialize("localhost")

		Convey("Test Message", func() {
			p := mockPost()
			_ = p.Save()

			c := mockComment(tmp_post_id_1, id2)
			c.PostId = p.Id.Hex()
			_ = c.Save()

			t.Logf("%s %s\n", p.Id.Hex(), c.PostId)

			um := mockMessage(c)

			err := um.Insert()
			So(err, ShouldBeNil)

			rm := mockMessage(c)
			rm.IsRead = true

			err = rm.Insert()
			So(err, ShouldBeNil)

			Convey("Get UnreadMessages", func() {
				messages := new(Messages)
				messages.GetUnreadMessages()

				So(messages, ShouldHaveLength, 1)
				So(messages.Get(0).Type, ShouldEqual, um.Type)
				So(messages.Get(0).Data, ShouldEqual, um.Data)
			})
		})
		Reset(func() {
			DropDatabase()
		})
	})
}
