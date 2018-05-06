package model

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/covrom/dingo/app/utils"
	"github.com/dinever/golf"
)

// const stmtSave = `INSERT OR REPLACE INTO tokens (id,value, user_id, created_at, expired_at) VALUES (?,?, ?, ?, ?)`
// const stmtGetTokenByValue = `SELECT * FROM tokens WHERE value = ?`

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
	t.UserId = string(u.Id)
	t.CreatedAt = utils.Now()
	expiredAt := t.CreatedAt.Add(time.Duration(expire) * time.Second)
	t.ExpiredAt = &expiredAt
	t.Value = utils.Sha1(fmt.Sprintf("%s-%s-%d-%s", ctx.ClientIP(), ctx.Request.UserAgent(), t.CreatedAt.Unix(), t.UserId))
	return t
}

// Save saves a token in the DB.
func (t *Token) Save() error {

	// session := mdb.Copy()
	// defer session.Close()

	if len(t.Id) == 0 {
		t.Id = bson.NewObjectId()
	}
	_, err := userSession.Clone().DB(DBName).C("tokens").UpsertId(t.Id, t)

	// // NOTE: since medder.Save doesn't support UNIQUE field, it is different from INSERT OR REPLACE...
	// // err := meddler.Save(db, "tokens", t) doens't work...
	// writeDB, err := db.Begin()
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	// _, err = writeDB.Exec(stmtSave, t.Id, t.Value, t.UserId, t.CreatedAt, t.ExpiredAt)
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	return err //writeDB.Commit()
}

// GetTokenByValue gets a token from the DB based on it's value.
func (t *Token) GetTokenByValue() error {
	// session := mdb.Copy()
	// defer session.Close()
	err := userSession.Clone().DB(DBName).C("tokens").Find(bson.M{"value": t.Value}).One(t)

	// err := meddler.QueryRow(db, t, stmtGetTokenByValue, t.Value)
	return err
}

// IsValid checks whether or not the token is valid.
func (t *Token) IsValid() bool {
	u := &User{Id: bson.ObjectId(t.UserId)}
	err := u.GetUserById()
	if err != nil {
		return false
	}
	return t.ExpiredAt.After(*utils.Now())
}
