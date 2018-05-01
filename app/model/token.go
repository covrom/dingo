package model

import (
	"fmt"
	"time"

	"github.com/dinever/golf"
	"github.com/covrom/dingo/app/utils"
	"github.com/russross/meddler"
)

const stmtSave = `INSERT OR REPLACE INTO tokens (id,value, user_id, created_at, expired_at) VALUES (?,?, ?, ?, ?)`
const stmtGetTokenByValue = `SELECT * FROM tokens WHERE value = ?`

// A Token is used to associate a user with a session.
type Token struct {
	Id        int64      `meddler:"id,pk"`
	Value     string     `meddler:"value"`
	UserId    int64      `meddler:"user_id"`
	CreatedAt *time.Time `meddler:"created_at"`
	ExpiredAt *time.Time `meddler:"expired_at"`
}

// NewToken creates a new token from the given user. Expire is the amount of
// time in seconds until expiry.
func NewToken(u *User, ctx *golf.Context, expire int64) *Token {
	t := new(Token)
	t.UserId = u.Id
	t.CreatedAt = utils.Now()
	expiredAt := t.CreatedAt.Add(time.Duration(expire) * time.Second)
	t.ExpiredAt = &expiredAt
	t.Value = utils.Sha1(fmt.Sprintf("%s-%s-%d-%d", ctx.ClientIP(), ctx.Request.UserAgent(), t.CreatedAt.Unix(), t.UserId))
	return t
}

// Save saves a token in the DB.
func (t *Token) Save() error {
	// NOTE: since medder.Save doesn't support UNIQUE field, it is different from INSERT OR REPLACE...
	// err := meddler.Save(db, "tokens", t) doens't work...
	writeDB, err := db.Begin()
	if err != nil {
		writeDB.Rollback()
		return err
	}
	_, err = writeDB.Exec(stmtSave, t.Id, t.Value, t.UserId, t.CreatedAt, t.ExpiredAt)
	if err != nil {
		writeDB.Rollback()
		return err
	}
	return writeDB.Commit()
}

// GetTokenByValue gets a token from the DB based on it's value.
func (t *Token) GetTokenByValue() error {
	err := meddler.QueryRow(db, t, stmtGetTokenByValue, t.Value)
	return err
}

// IsValid checks whether or not the token is valid.
func (t *Token) IsValid() bool {
	u := &User{Id: t.UserId}
	err := u.GetUserById()
	if err != nil {
		return false
	}
	return t.ExpiredAt.After(*utils.Now())
}
