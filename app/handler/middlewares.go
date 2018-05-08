package handler

import (
	"net/http"

	"github.com/globalsign/mgo/bson"

	"github.com/covrom/dingo/app/model"
	"github.com/dinever/golf"
)

func AuthMiddleware(next golf.HandlerFunc) golf.HandlerFunc {
	fn := func(ctx *golf.Context) {
		userNum, err := model.GetNumberOfUsers()
		if err == nil && userNum == 0 {
			ctx.Redirect("/signup/")
			return
		}
		tokenStr, err := ctx.Request.Cookie("token-value")
		if err != nil {
			ctx.Redirect("/login/")
			return
		}
		//token, err := model.GetTokenByValue(tokenStr.Value)
		token := &model.Token{Value: tokenStr.Value}
		err = token.GetTokenByValue()
		if err != nil || !token.IsValid() {
			ctx.Redirect("/login/")
			return
		}
		tokenUser, err := ctx.Request.Cookie("token-user")
		if err != nil {
			ctx.Redirect("/login/")
			return
		}
		uid := tokenUser.Value
		user := &model.User{Id: bson.ObjectIdHex(uid)}
		err = user.GetUserById()
		if err != nil {
			panic(err)
		}
		ctx.Session.Set("user", user)
		next(ctx)
	}
	return fn
}

func JWTAuthMiddleware(next golf.HandlerFunc) golf.HandlerFunc {
	return func(ctx *golf.Context) {
		tokenHeader := ctx.Header("X-SESSION-TOKEN")
		if tokenHeader == "" {
			ctx.SendStatus(http.StatusUnauthorized)
			return
		}
		token, err := model.ValidateJWT(tokenHeader)
		if err != nil {
			ctx.SendStatus(http.StatusUnauthorized)
			return
		}
		ctx.Session.Set("jwt", model.NewJWTFromToken(token))
		next(ctx)
	}
}
