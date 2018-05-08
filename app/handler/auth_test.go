package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dinever/golf"
	"github.com/covrom/dingo/app/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUserSignUpWithCorrectInformation(t *testing.T) {
	Convey("Initialize database", t, func() {
		testDB := fmt.Sprintf(filepath.Join(os.TempDir(), "ding-testdb-%s"), fmt.Sprintf(time.Now().Format("20060102T150405.000")))
		model.Initialize(testDB, true)

		Convey("When the form data is correct", func() {
			ctx := mockSignUpPostRequest("johndoe@email.com", "johndoe", "somepassword", "somepassword")

			Convey("The request is handled by the signup handler", func() {
				AuthSignUpHandler(ctx)

				Convey("The response code should be 200", func() {
					So(ctx.StatusCode(), ShouldEqual, 200)
				})

				Convey("Get the previously signed user", func() {
					user := &model.User{Id: model.Tmp_id_1}
					err := user.GetUserById()

					Convey("The user can be get without error", func() {
						So(err, ShouldBeNil)
					})

					Convey("The user should not be nil", func() {
						So(user, ShouldNotBeNil)
					})
				})

				Convey("The response code should be", func() {
					recorder := ctx.Response.(*httptest.ResponseRecorder)
					request := &http.Request{Header: http.Header{"Cookie": recorder.HeaderMap["Set-Cookie"]}}
					cookie, _ := request.Cookie("token-value")
					So(cookie.Value, ShouldNotBeEmpty)
				})
			})
		})

		Reset(func() {
			os.Remove(testDB)
		})
	})
}

func TestUserSignUpWithIncorrectInformation(t *testing.T) {
	Convey("Initialize database", t, func() {
		testDB := fmt.Sprintf(filepath.Join(os.TempDir(), "ding-testdb-%s"), fmt.Sprintf(time.Now().Format("20060102T150405.000")))
		model.Initialize(testDB, true)

		Convey("When the email is illegal", func() {

			ctx := mockSignUpPostRequest("johndoe.com", "johndoe", "somepassword", "somepassword")

			Convey("Handler should raise a 400 error", func() {
				AuthSignUpHandler(ctx)
				So(ctx.StatusCode(), ShouldEqual, 400)
			})

		})

		Convey("When the password is too short", func() {

			ctx := mockSignUpPostRequest("johndoe@gmail.com", "johndoe", "some", "some")

			Convey("Handler should raise a 400 error", func() {
				AuthSignUpHandler(ctx)
				So(ctx.StatusCode(), ShouldEqual, 400)
			})
		})

		Convey("When the password is too long", func() {

			ctx := mockSignUpPostRequest("johndoe@gmail.com", "johndoe", "somepasswordistoolongforsomebody", "somepasswordistoolongforsomebody")

			Convey("Handler should raise a 400 error", func() {
				AuthSignUpHandler(ctx)
				So(ctx.StatusCode(), ShouldEqual, 400)
			})
		})

		Convey("When the password does not match", func() {

			ctx := mockSignUpPostRequest("johndoe@gmail.com", "johndoe", "somepassword", "someanotherpassword")

			Convey("Handler should raise a 400 error", func() {
				AuthSignUpHandler(ctx)
				So(ctx.StatusCode(), ShouldEqual, 400)
			})
		})

		Convey("When the name is empty", func() {
			ctx := mockSignUpPostRequest("johndoe@gmail.com", "", "somepassword", "somepassword")

			Convey("Handler should raise a 400 error", func() {
				AuthSignUpHandler(ctx)
				So(ctx.StatusCode(), ShouldEqual, 400)
			})
		})

		Convey("When the password is empty", func() {
			ctx := mockSignUpPostRequest("johndoe@gmail.com", "johndoe", "", "")

			Convey("Handler should raise a 400 error", func() {
				AuthSignUpHandler(ctx)
				So(ctx.StatusCode(), ShouldEqual, 400)
			})
		})

		Reset(func() {
			os.Remove(testDB)
		})
	})
}

func TestUserLogInWithCorrectInformation(t *testing.T) {
	Convey("Initialize database", t, func() {
		testDB := fmt.Sprintf(filepath.Join(os.TempDir(), "ding-testdb-%s"), fmt.Sprintf(time.Now().Format("20060102T150405.000")))
		model.Initialize(testDB, true)

		Convey("Create a user first", func() {
			err := model.NewUser(email, name).Create(password)
			So(err, ShouldBeNil)

			Convey("Login with correct information", func() {

				ctx := mockLogInPostContext()
				rec := ctx.Response.(*httptest.ResponseRecorder)
				AuthLoginHandler(ctx)

				tokenValue := retrieveSetCookieValue(rec, "token-value")

				Convey("Should set token value correctly", func() {
					So(len(tokenValue), ShouldEqual, 40)
				})

				Convey("Should return json correctly", func() {
					So(rec.Body.String(), ShouldContainSubstring, "success")
				})

				Convey("Visit admin dashboard", func() {

					w := httptest.NewRecorder()
					app := golf.New()
					app = InitTestApp(app)

					req := makeTestHTTPRequest(nil, "GET", "/admin/")
					req.Header = http.Header{"Cookie": rec.HeaderMap["Set-Cookie"]}
					ctx = golf.NewContext(req, w, app)

					ctx.App.ServeHTTP(ctx.Response, ctx.Request)

					Convey("Should not redirect to login page ", func() {
						So(ctx.Response.(*httptest.ResponseRecorder).Code, ShouldEqual, 200)
					})
				})
			})
		})

		Reset(func() {
			os.Remove(testDB)
		})
	})
}

func TestUserWithoutAuthentication(t *testing.T) {

	Convey("Initialize database", t, func() {
		testDB := fmt.Sprintf(filepath.Join(os.TempDir(), "ding-testdb-%s"), fmt.Sprintf(time.Now().Format("20060102T150405.000")))
		model.Initialize(testDB, true)

		Convey("Visit admin dashbaord", func() {
			ctx := mockContext(nil, "GET", "/admin/")
			ctx.App.ServeHTTP(ctx.Response, ctx.Request)

			Convey("Status code should be 302 redirection", func() {
				So(ctx.Response.(*httptest.ResponseRecorder).Code, ShouldEqual, 302)
			})
		})

		Reset(func() {
			os.Remove(testDB)
		})
	})
}
