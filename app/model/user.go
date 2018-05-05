package model

import (
	"time"

	"github.com/covrom/dingo/app/utils"
	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
)

// const stmtGetUserById = `SELECT * FROM users WHERE id = ?`
// const stmtGetUserBySlug = `SELECT * FROM users WHERE slug = ?`
// const stmtGetUserByName = `SELECT * FROM users WHERE name = ?`
// const stmtGetUserByEmail = `SELECT * FROM users WHERE email = ?`
// const stmtInsertRoleUser = `INSERT INTO roles_users (id, role_id, user_id) VALUES (?, ?, ?)`
// const stmtGetUsersCountByEmail = `SELECT count(*) FROM users where email = ?`
// const stmtGetNumberOfUsers = `SELECT COUNT(*) FROM users`

// A User is a user on the site.
type User struct {
	Id             bson.ObjectId `json:"_id"`
	Name           string        `json:"name"`
	Slug           string        `json:"slug"`
	HashedPassword string        `json:"password"`
	Email          string        `json:"email"`
	Image          string        `json:"image"`    // NULL
	Cover          string        `json:"cover"`    // NULL
	Bio            string        `json:"bio"`      // NULL
	Website        string        `json:"website"`  // NULL
	Location       string        `json:"location"` // NULL
	Accessibility  string        `json:"accessibility"`
	Status         string        `json:"status"`
	Language       string        `json:"language"`
	Lastlogin      *time.Time    `json:"last_login"`
	CreatedAt      *time.Time    `json:"created_at"`
	CreatedBy      string        `json:"created_by"`
	UpdatedAt      *time.Time    `json:"updated_at"`
	UpdatedBy      string        `json:"updated_by"`
	Role           int           `json:"-"` //1 = Administrator, 2 = Editor, 3 = Author, 4 = Owner
}

var ghostUser = &User{Id: "", Name: "Blog User", Email: "example@example.com"}

// NewUser creates a new user from the given email and name, with the CreatedAt
// and UpdatedAt fields set to the current time.
func NewUser(email, name string) *User {
	return &User{
		Id:        bson.NewObjectId(),
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
	u.CreatedBy = ""
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
	session := mdb.Copy()
	defer session.Close()
	if len(u.Id) == 0 {
		u.Id = bson.NewObjectId()
	}
	if len(u.Slug) == 0 {
		u.Slug = GenerateSlug(string(u.Id)+u.Email, "users")
	}
	_, err := session.DB(DBName).C("users").UpsertId(u.Id, u)

	// err := meddler.Update(db, "users", u)
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
	session := mdb.Copy()
	defer session.Close()
	err := session.DB(DBName).C("users").FindId(u.Id).One(u)

	// err := meddler.QueryRow(db, u, stmtGetUserById, u.Id)
	return err
}

// GetUserBySlug finds the user by their slug in the DB.
func (u *User) GetUserBySlug() error {
	session := mdb.Copy()
	defer session.Close()

	err := session.DB(DBName).C("users").Find(bson.M{"slug": u.Slug}).One(u)
	// err := meddler.QueryRow(db, u, stmtGetUserBySlug, u.Slug)
	return err
}

// GetUserByName finds the user by name in the DB.
func (u *User) GetUserByName() error {
	session := mdb.Copy()
	defer session.Close()
	err := session.DB(DBName).C("users").Find(bson.M{"name": u.Name}).One(u)

	// err := meddler.QueryRow(db, u, stmtGetUserByName, u.Name)
	return err
}

// GetUserByEmail finds the user by email in the DB.
func (u *User) GetUserByEmail() error {
	session := mdb.Copy()
	defer session.Close()
	err := session.DB(DBName).C("users").Find(bson.M{"email": u.Email}).One(u)

	// err := meddler.QueryRow(db, u, stmtGetUserByEmail, u.Email)
	return err
}

// Insert inserts the user into the DB.
func (u *User) Insert() error {
	session := mdb.Copy()
	defer session.Close()
	if len(u.Id) == 0 {
		u.Id = bson.NewObjectId()
	}
	if len(u.Slug) == 0 {
		u.Slug = GenerateSlug(string(u.Id)+u.Email, "users")
	}
	_, err := session.DB(DBName).C("users").UpsertId(u.Id, u)

	// err := meddler.Insert(db, "users", u)
	return err
}

type RolesUsers struct {
	RoleId string `json:"role_id"`
	UserId string `json:"user_id"`
}

// InsertRoleUser assigns a role to the given user based on the given Role ID.
func InsertRoleUser(role_id string, user_id string) error {
	// writeDB, err := db.Begin()
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	session := mdb.Copy()
	defer session.Close()
	err := session.DB(DBName).C("roles_users").Insert(&RolesUsers{RoleId: role_id, UserId: user_id})

	// _, err = writeDB.Exec(stmtInsertRoleUser, nil, role_id, user_id)
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	return err //writeDB.Commit()
}

// UserEmailExist checks to see if the given User's email exists.
func (u User) UserEmailExist() bool {
	session := mdb.Copy()
	defer session.Close()
	count, err := session.DB(DBName).C("users").Find(bson.M{"email": u.Email}).Count()

	// var count int64
	// row := db.QueryRow(stmtGetUsersCountByEmail, u.Email)
	// err := row.Scan(&count)
	if count > 0 || err != nil {
		return true
	}
	return false
}

// GetNumberOfUsers returns the total number of users.
func GetNumberOfUsers() (int64, error) {
	session := mdb.Copy()
	defer session.Close()
	count, err := session.DB(DBName).C("users").Find(bson.M{}).Count()

	// var count int64
	// row := db.QueryRow(stmtGetNumberOfUsers)
	// err := row.Scan(&count)
	return int64(count), err
}
