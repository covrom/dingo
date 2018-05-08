package model

import (
	// "database/sql"

	"time"

	"fmt"

	"github.com/covrom/dingo/app/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	// "github.com/russross/meddler"
)

// Comments are a slice of "Comment"s
type Comments []*Comment

// A Comment defines comment item data.
type Comment struct {
	Id        bson.ObjectId
	PostId    string
	Author    string
	Email     string
	Avatar    string
	Website   string
	Ip        string
	CreatedAt *time.Time
	Content   string
	Approved  bool
	UserAgent string
	Type      string
	Parent    string
	UserId    string
	Children  *Comments `json:"-" bson:"-"`
}

// Len returns the number of "Comment"s in a "Comments".
func (c Comments) Len() int {
	return len(c)
}

// Get returns the Comment at the given index.
func (c Comments) Get(i int) *Comment {
	return c[i]
}

// GetAll returns a slice of all the "Comment"s.
func (c Comments) GetAll() []*Comment {
	return c
}

// NewComment returns a new comment, with the CreatedAt field set to the
// current time.
func NewComment() *Comment {
	return &Comment{
		Id:        bson.NewObjectId(),
		CreatedAt: utils.Now(),
	}
}

// Save saves the comment in the DB.
func (c *Comment) Save() error {
	c.Avatar = utils.Gravatar(c.Email, "50")

	if len(c.Id) == 0 {
		c.Id = bson.NewObjectId()
	}
	_, err := comSession.Clone().DB(DBName).C("comments").UpsertId(c.Id, c)
	return err

	// err := meddler.Save(db, "comments", c)
	// return err
}

// ToJson returns a comment as a map, in order to be encoded as JSON.
func (c *Comment) ToJson() map[string]interface{} {
	m := make(map[string]interface{})
	m["id"] = c.Id
	m["author"] = c.Author
	m["email"] = c.Email
	m["website"] = c.Website
	m["avatar"] = c.Avatar
	m["content"] = c.Content
	m["create_time"] = c.CreatedAt.Unix()
	m["pid"] = c.Parent
	m["approved"] = c.Approved
	m["ip"] = c.Ip
	m["user_agent"] = c.UserAgent
	m["parent_content"] = c.ParentContent()
	return m
}

// ParentContent returns the parent of a given comment, if it exists. Used for
// threaded comments.
func (c *Comment) ParentContent() string {
	if len(c.Parent) == 0 {
		return ""
	}

	comment := &Comment{Id: bson.ObjectIdHex(c.Parent)}
	err := comment.GetCommentById()
	if err != nil {
		return "> Comment not found."
	}
	str := "> @" + comment.Author + "\n\n"
	str += "> " + comment.Content + "\n"
	return str
}

// GetNumberOfComments returns the total number of comments in the DB.
func GetNumberOfComments() (int64, error) {

	// session := mdb.Copy()
	// defer session.Close()
	count, err := comSession.Clone().DB(DBName).C("comments").Count()

	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// GetCommentList returns a new pager based on the total number of comments.
func (c *Comments) GetCommentList(page, size int64, onlyApproved bool) (*utils.Pager, error) {
	var pager *utils.Pager

	count, err := GetNumberOfComments()
	pager = utils.NewPager(page, size, count)

	if !pager.IsValid {
		return pager, fmt.Errorf("Page not found")
	}

	// session := mdb.Copy()
	// defer session.Close()

	if onlyApproved {
		err = comSession.Clone().DB(DBName).C("comments").Find(bson.M{"approved": true}).Sort("-createdat").Skip(int(pager.Begin)).Limit(int(size)).All(c)
	} else {
		err = comSession.Clone().DB(DBName).C("comments").Find(bson.M{}).Sort("-createdat").Skip(int(pager.Begin)).Limit(int(size)).All(c)
	}

	// err = meddler.QueryAll(db, c, fmt.Sprintf(stmtGetCommentList, where), size, pager.Begin)
	return pager, err
}

// GetCommentById gets a comment by its ID, and populates that comment struct
// with the contents for that comment from the DB.
func (c *Comment) GetCommentById() error {
	// session := mdb.Copy()
	// defer session.Close()

	err := comSession.Clone().DB(DBName).C("comments").FindId(c.Id).One(c)

	// err := meddler.QueryRow(db, c, stmtGetCommentById, c.Id)
	return err
}

func (c *Comment) getChildComments() (*Comments, error) {
	// session := mdb.Copy()
	// defer session.Close()

	comments := new(Comments)
	err := comSession.Clone().DB(DBName).C("comments").Find(bson.M{"parent": c.Id.Hex(), "approved": true}).All(comments)

	// err := meddler.QueryAll(db, comments, stmtGetCommentsByParentId, c.Id)
	return comments, err
}

// ParentComment returns the associated parent Comment, if one exists.
func (c *Comment) ParentComment() (*Comment, error) {
	parent := NewComment()
	parent.Id = bson.ObjectIdHex(c.Parent)
	return parent, parent.GetCommentById()
}

// Post returns the post associated with the commment.
func (c *Comment) Post() *Post {
	post := NewPost()
	post.Id = bson.ObjectIdHex(c.PostId)
	post.GetPostById()
	return post
}

// GetCommentsByPostId gets all the comments for the given post ID.
func (comments *Comments) GetCommentsByPostId(id string) error {
	// session := mdb.Copy()
	// defer session.Close()

	err := comSession.Clone().DB(DBName).C("comments").Find(bson.M{"postid": id, "parent": "", "approved": true}).All(comments)

	// err := meddler.QueryAll(db, comments, stmtGetParentCommentsByPostId, id)
	for _, c := range *comments {
		buildCommentTree(c, c, 1)
	}
	return err

}

func buildCommentTree(p *Comment, c *Comment, level int) {
	childComments, _ := c.getChildComments()
	if p.Children == nil {
		p.Children = childComments
	} else {
		newChildComments := append(*p.Children, *childComments...)
		p.Children = &newChildComments
	}
	for _, c := range *childComments {
		if level >= 2 {
			buildCommentTree(p, c, level+1)
		} else {
			buildCommentTree(c, c, level+1)
		}
	}
}

// DeleteComment deletes the comment with the given ID from the DB.
func DeleteComment(id string) error {
	// session := mdb.Copy()
	// defer session.Close()
	session := comSession.Clone()

	childs := new(Comments)
	err := session.DB(DBName).C("comments").Find(bson.M{"parent": id}).All(childs)
	if err == nil {
		for _, child := range *childs {
			if len(child.Id) > 0 {
				DeleteComment(child.Id.Hex())
			}
		}
	}

	err = session.DB(DBName).C("comments").RemoveId(bson.ObjectIdHex(id))
	if err == mgo.ErrNotFound {
		err = nil
	}

	// writeDB, err := db.Begin()
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	// _, err = writeDB.Exec(stmtDeleteCommentById, id)
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	return err //writeDB.Commit()
}

// ValidateComment validates a comment to ensure that all required data exists
// and is valid. Returns an empty string on success.
func (c *Comment) ValidateComment() string {
	if utils.IsEmptyString(c.Author) || utils.IsEmptyString(c.Content) {
		return "Name, Email and Content are required fields."
	}
	if !utils.IsEmail(c.Email) {
		return "Email format not valid."
	}
	if !utils.IsEmptyString(c.Website) && !utils.IsURL(c.Website) {
		return "Website URL format not valid."
	}
	return ""
}

// const stmtGetAllCommentCount = `SELECT count(*) FROM comments`
// const stmtDeleteCommentById = `DELETE FROM comments WHERE id = ?`
// const stmtGetCommentList = `SELECT * FROM comments %s ORDER BY created_at DESC LIMIT ? OFFSET ?`
// const stmtGetCommentById = `SELECT * FROM comments WHERE id = ?`
// const stmtGetCommentsByPostId = `SELECT * FROM comments WHERE post_id = ? AND approved = 1 AND parent = 0`
// const stmtGetParentCommentsByPostId = `SELECT * FROM comments WHERE post_id = ? AND approved = 1 AND parent = 0`
// const stmtGetCommentsByParentId = `SELECT * FROM comments WHERE parent = ? AND approved = 1`
