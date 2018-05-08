package handler

import (
	"github.com/globalsign/mgo/bson"

	"github.com/covrom/dingo/app/model"
	"github.com/dinever/golf"
)

func registerUserHandlers(app *golf.Application, routes map[string]map[string]interface{}) {
	app.Get("/api/users", APIUsersHandler)
	routes["GET"]["users_url"] = "/api/users"

	app.Get("/api/users/:user_id", APIUserHandler)
	routes["GET"]["user_url"] = "/api/users/:user_id"

	app.Get("/api/users/slug/:slug", APIUserSlugHandler)
	routes["GET"]["user_slug_url"] = "/api/users/slug/:slug"

	app.Get("/api/users/email/:email", APIUserEmailHandler)
	routes["GET"]["user_email_url"] = "/api/users/email/:email"
}

// APIUserHandler retrieves the user with the given id.
func APIUserHandler(ctx *golf.Context) {
	id := ctx.Param("user_id")
	// if err != nil {
	// 	handleErr(ctx, 500, err)
	// 	return
	// }
	user := &model.User{Id: bson.ObjectIdHex(id)}
	err := user.GetUserById()
	if err != nil {
		handleErr(ctx, 404, err)
		return
	}
	ctx.JSONIndent(user, "", "  ")
}

// APIUserSlugHandler retrives the user with the given slug.
func APIUserSlugHandler(ctx *golf.Context) {
	slug := ctx.Param("slug")
	user := &model.User{Slug: slug}
	err := user.GetUserBySlug()
	if err != nil {
		handleErr(ctx, 404, err)
		return
	}
	ctx.JSONIndent(user, "", "  ")
}

// APIUserEmailHandler retrieves the user with the given email.
func APIUserEmailHandler(ctx *golf.Context) {
	email := ctx.Param("email")
	user := &model.User{Email: email}
	err := user.GetUserByEmail()
	if err != nil {
		handleErr(ctx, 404, err)
		return
	}
	ctx.JSONIndent(user, "", "  ")
}

// APIUsersHandler retrieves all users.
func APIUsersHandler(ctx *golf.Context) {
	ctx.JSONIndent(map[string]interface{}{
		"message": "Not implemented",
	}, "", "  ")
}
