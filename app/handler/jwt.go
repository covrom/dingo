package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dinever/golf"
	"github.com/covrom/dingo/app/model"
)

type JWTPostBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func registerJWTHandlers(app *golf.Application, routes map[string]map[string]interface{}) {
	adminChain := golf.NewChain(JWTAuthMiddleware)
	// Auth
	app.Post("/auth", JWTAuthLoginHandler)
	routes["POST"]["auth_new_url"] = "/auth"

	app.Get("/auth", adminChain.Final(JWTDecryptHandler))
	routes["GET"]["auth_decrypt_url"] = "/auth"
}

func JWTAuthLoginHandler(ctx *golf.Context) {
	var email string
	var password string
	contentType := ctx.Header("Content-Type")
	ctx.SetHeader("Content-Type", "application/json")
	switch {
	case strings.Contains(contentType, "application/json"):
		defer ctx.Request.Body.Close()
		body, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.SendStatus(http.StatusInternalServerError)
			ctx.JSON(map[string]interface{}{"status": "error: unable to read request body"})
			return
		}
		var unmarshalledRequestBody JWTPostBody
		json.Unmarshal(body, &unmarshalledRequestBody)
		email = unmarshalledRequestBody.Email
		password = unmarshalledRequestBody.Password
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		email = ctx.Request.FormValue("email")
		password = ctx.Request.FormValue("password")
	default:
		ctx.SendStatus(http.StatusBadRequest)
		ctx.JSON(map[string]interface{}{"status": "error: unrecognized Content-Type"})
		return
	}
	user := &model.User{Email: email}
	err := user.GetUserByEmail()
	if user == nil || err != nil {
		ctx.SendStatus(http.StatusUnauthorized)
		ctx.JSON(map[string]interface{}{"status": "error"})
		return
	}
	if !user.CheckPassword(password) {
		ctx.SendStatus(http.StatusUnauthorized)
		ctx.JSON(map[string]interface{}{"status": "error"})
		return
	}

	token, err := model.NewJWT(user)
	if err != nil {
		log.Printf("Unable to sign token: %v\n", err)
		ctx.SendStatus(http.StatusInternalServerError)
		ctx.JSON(map[string]interface{}{"status": "error"})
		return
	}

	ctx.SendStatus(http.StatusOK)
	ctx.JSON(token)
}

func JWTDecryptHandler(ctx *golf.Context) {
	if ctx.Header("Content-Type") != "application/json" {
		ctx.SetHeader("Content-Type", "application/json")
	}
	token, err := ctx.Session.Get("jwt")
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		return
	}
	ctx.SendStatus(http.StatusOK)
	ctx.JSON(token)
}
