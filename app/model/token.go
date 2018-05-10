package model

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/covrom/dingo/app/utils"
	"github.com/dinever/golf"
)

// A Token is used to associate a user with a session.
type Token struct {
	Id        bson.ObjectId `bson:"_id"`
	Value     string
	UserId    string
	CreatedAt *time.Time
	ExpiredAt *time.Time
}

// NewToken creates a new token from the given user. Expire is the amount of
// time in seconds until expiry.
func NewToken(u *User, ctx *golf.Context, expire int64) *Token {
	t := new(Token)
	t.UserId = u.Id.Hex()
	t.CreatedAt = utils.Now()
	expiredAt := t.CreatedAt.Add(time.Duration(expire) * time.Second)
	t.ExpiredAt = &expiredAt
	t.Value = utils.Sha1(fmt.Sprintf("%s-%s-%d-%s", ctx.ClientIP(), ctx.Request.UserAgent(), t.CreatedAt.Unix(), t.UserId))
	return t
}

// Save saves a token in the DB.
func (t *Token) Save() error {

	if len(t.Id) == 0 {
		t.Id = bson.NewObjectId()
	}
	_, err := userSession.Clone().DB(DBName).C("tokens").UpsertId(t.Id, t)

	return err
}

// GetTokenByValue gets a token from the DB based on it's value.
func (t *Token) GetTokenByValue() error {
	err := userSession.Clone().DB(DBName).C("tokens").Find(bson.M{"value": t.Value}).One(t)
	return err
}

// IsValid checks whether or not the token is valid.
func (t *Token) IsValid() bool {
	u := &User{Id: bson.ObjectIdHex(t.UserId)}
	err := u.GetUserById()
	if err != nil {
		return false
	}
	return t.ExpiredAt.After(*utils.Now())
}
