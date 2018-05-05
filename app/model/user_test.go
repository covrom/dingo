package model

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const name = "Shawn Ding"
const email = "dingpeixuan911@gmail.com"
const password = "passwordfortest"

func mockUser() *User {
	u := NewUser(email, name)
	return u
}

func userEqualCheck(user *User, expected *User) {
	So(user.Id, ShouldEqual, expected.Id)
	So(user.Name, ShouldEqual, expected.Name)
	So(user.Slug, ShouldNotBeNil)
	So(user.Avatar, ShouldNotBeNil)
	So(user.Email, ShouldEqual, expected.Email)
	So(user.Role, ShouldNotBeNil)
}

func TestUser(t *testing.T) {
	Convey("Initialize database", t, func() {
		DBName = fmt.Sprintf("ding-testdb-%s", time.Now().Format("20060102T150405"))
		Initialize("localhost")

		Convey("Test User", func() {
			user := mockUser()
			err := user.Create(password)
			So(err, ShouldBeNil)

			Convey("Get User By Id", func() {
				u := &User{Id: user.Id}
				err := u.GetUserById()
				So(err, ShouldBeNil)
				userEqualCheck(u, user)
			})

			Convey("Get User By Slug", func() {
				u := &User{Slug: user.Slug}
				err := u.GetUserBySlug()
				So(err, ShouldBeNil)
				userEqualCheck(u, user)
			})

			Convey("Get User By Name", func() {
				u := &User{Name: user.Name}
				err := u.GetUserByName()
				So(err, ShouldBeNil)
				userEqualCheck(u, user)
			})

			Convey("Get User By Email", func() {
				u := &User{Email: user.Email}
				err := u.GetUserByEmail()
				So(err, ShouldBeNil)
				userEqualCheck(u, user)
			})

			Convey("Check Password", func() {
				result := user.CheckPassword("passwordfortest")
				So(result, ShouldEqual, true)
			})

			Convey("Change Password", func() {
				err := user.ChangePassword("updatedpassword")
				So(err, ShouldBeNil)
				result := user.CheckPassword("updatedpassword")
				So(result, ShouldEqual, true)
			})

			Convey("Email Exist", func() {
				result := user.UserEmailExist()
				So(err, ShouldBeNil)
				So(result, ShouldEqual, true)
			})

			Convey("Get Number Of Users", func() {
				result, err := GetNumberOfUsers()
				So(err, ShouldBeNil)
				So(result, ShouldEqual, 1)
			})

			Convey("Get Avatar", func() {
				u := &User{Id: user.Id}
				err := u.GetUserById()
				So(err, ShouldBeNil)
				So(u.Avatar(), ShouldEqual, "http://1.gravatar.com/avatar/3583d6fbf01855a8d637059044752eb8?s=150")
			})

			Convey("Update User", func() {
				user.Name = "Kenjiro Nakayama"
				user.Email = "nakayamakenjiro@gmail.com"
				err := user.Update()
				So(err, ShouldBeNil)
				u := &User{Id: user.Id}
				err = u.GetUserById()
				userEqualCheck(u, user)
			})

		})
		Reset(func() {
			DropDatabase()
		})
	})
}
