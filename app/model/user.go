package model

import (
	"time"

	"github.com/covrom/dingo/app/utils"
	"github.com/russross/meddler"
	"golang.org/x/crypto/bcrypt"
)

const stmtGetUserById = `SELECT * FROM users WHERE id = ?`
const stmtGetUserBySlug = `SELECT * FROM users WHERE slug = ?`
const stmtGetUserByName = `SELECT * FROM users WHERE name = ?`
const stmtGetUserByEmail = `SELECT * FROM users WHERE email = ?`
const stmtInsertRoleUser = `INSERT INTO roles_users (id, role_id, user_id) VALUES (?, ?, ?)`
const stmtGetUsersCountByEmail = `SELECT count(*) FROM users where email = ?`
const stmtGetNumberOfUsers = `SELECT COUNT(*) FROM users`

// A User is a user on the site.
type User struct {
	Id             int64      `meddler:"id,pk"`
	Name           string     `meddler:"name"`
	Slug           string     `meddler:"slug"`
	HashedPassword string     `meddler:"password"`
	Email          string     `meddler:"email"`
	Image          string     `meddler:"image"`    // NULL
	Cover          string     `meddler:"cover"`    // NULL
	Bio            string     `meddler:"bio"`      // NULL
	Website        string     `meddler:"website"`  // NULL
	Location       string     `meddler:"location"` // NULL
	Accessibility  string     `meddler:"accessibility"`
	Status         string     `meddler:"status"`
	Language       string     `meddler:"language"`
	Lastlogin      *time.Time `meddler:"last_login"`
	CreatedAt      *time.Time `meddler:"created_at"`
	CreatedBy      int        `meddler:"created_by"`
	UpdatedAt      *time.Time `meddler:"updated_at"`
	UpdatedBy      int        `meddler:"updated_by"`
	Role           int        `meddler:"-"` //1 = Administrator, 2 = Editor, 3 = Author, 4 = Owner
}

var ghostUser = &User{Id: 0, Name: "Dingo User", Email: "example@example.com"}

// NewUser creates a new user from the given email and name, with the CreatedAt
// and UpdatedAt fields set to the current time.
func NewUser(email, name string) *User {
	return &User{
		Email:     email,
		Name:      name,
		CreatedAt: utils.Now(),
		UpdatedAt: utils.Now(),
	}
}

// Create saves a user in the DB with the given password, first hashing and
// salting that password via bcrypt.
func (u *User) Create(password string) error {
	var err error
	u.HashedPassword, err = EncryptPassword(password)
	if err != nil {
		return err
	}
	u.CreatedBy = 0
	return u.Save()
}

// Save saves a user to the DB.
func (u *User) Save() error {
	err := u.Insert()
	//	err = InsertRoleUser(u.Role, userId)
	//	if err != nil {
	//		return err
	//	}
	return err
}

// Update updates an existing user in the DB.
func (u *User) Update() error {
	u.UpdatedAt = utils.Now()
	// TODO:
	//u.UpdatedBy = ...
	err := meddler.Update(db, "users", u)
	return err
}

// ChangePassword changes the password for the given user.
func (u *User) ChangePassword(password string) error {
	var err error
	u.HashedPassword, err = EncryptPassword(password)
	if err != nil {
		return err
	}
	err = u.Update()
	return err
}

// EncrypPassword hashes and salts the given password via bcrypt, returning
// the newly hashed and salted password.
func EncryptPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword checks to see if the given password matches the hashed password
// for the given user, returning true if it's a match.
func (u *User) CheckPassword(password string) bool {
	err := u.GetUserByEmail()
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	if err != nil {
		return false
	}
	return true
}

// Avatar returns the Gravatar of the given user, with the Gravatar being
// 150px by 150px.
func (u *User) Avatar() string {
	return utils.Gravatar(u.Email, "150")
}

// GetUserById finds the user by ID in the DB.
func (u *User) GetUserById() error {
	err := meddler.QueryRow(db, u, stmtGetUserById, u.Id)
	return err
}

// GetUserBySlug finds the user by their slug in the DB.
func (u *User) GetUserBySlug() error {
	err := meddler.QueryRow(db, u, stmtGetUserBySlug, u.Slug)
	return err
}

// GetUserByName finds the user by name in the DB.
func (u *User) GetUserByName() error {
	err := meddler.QueryRow(db, u, stmtGetUserByName, u.Name)
	return err
}

// GetUserByEmail finds the user by email in the DB.
func (u *User) GetUserByEmail() error {
	err := meddler.QueryRow(db, u, stmtGetUserByEmail, u.Email)
	return err
}

// Insert inserts the user into the DB.
func (u *User) Insert() error {
	err := meddler.Insert(db, "users", u)
	return err
}

// InsertRoleUser assigns a role to the given user based on the given Role ID.
func InsertRoleUser(role_id int, user_id int64) error {
	writeDB, err := db.Begin()
	if err != nil {
		writeDB.Rollback()
		return err
	}
	_, err = writeDB.Exec(stmtInsertRoleUser, nil, role_id, user_id)
	if err != nil {
		writeDB.Rollback()
		return err
	}
	return writeDB.Commit()
}

// UserEmailExist checks to see if the given User's email exists.
func (u User) UserEmailExist() bool {
	var count int64
	row := db.QueryRow(stmtGetUsersCountByEmail, u.Email)
	err := row.Scan(&count)
	if count > 0 || err != nil {
		return true
	}
	return false
}

// GetNumberOfUsers returns the total number of users.
func GetNumberOfUsers() (int64, error) {
	var count int64
	row := db.QueryRow(stmtGetNumberOfUsers)
	err := row.Scan(&count)
	return count, err
}
