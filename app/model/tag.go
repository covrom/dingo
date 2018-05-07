package model

import (
	"log"
	"strings"
	"time"

	"github.com/covrom/dingo/app/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// A Tag is a keyword associated with a post.
type Tag struct {
	Id        bson.ObjectId `bson:"_id"`
	Name      string
	Slug      string
	Hidden    bool
	CreatedAt *time.Time
	CreatedBy string
	UpdatedAt *time.Time
	UpdatedBy string
}

// Url returns the URL of the given slug.
func (t *Tag) Url() string {
	return "/tag/" + t.Slug
}

// Tags are a slice of "Tag"s
type Tags []*Tag

// Len returns the amount of "Tag"s in the Tags slice.
func (t Tags) Len() int {
	return len(t)
}

// Get returns a Tag at the given index.
func (t Tags) Get(i int) *Tag {
	return t[i]
}

// GetAll returns a slice of every Tag.
func (t Tags) GetAll() []*Tag {
	return t
}

// NewTag creates a new Tag, with CreatedAt being set to the current time.
func NewTag(name, slug string) *Tag {
	return &Tag{
		Id:        bson.NewObjectId(),
		Name:      name,
		Slug:      slug,
		CreatedAt: utils.Now(),
	}
}

// Save saves a Tag to the DB.
func (t *Tag) Save() error {
	oldTag := &Tag{Slug: t.Slug}
	err := oldTag.GetTagBySlug()
	if err != nil && err == mgo.ErrNotFound {
		// fmt.Printf("TAG %v INSERTED!!!", *t)
		if err := t.Insert(); err != nil {
			log.Printf("[Error] Can not insert tag: %v", err.Error())
			return err
		}
	} else if err != nil {
		return err
	} else {
		t.Id = oldTag.Id
		// If oldTag.Hidden != t.Hidden, we need to decide whether change the hidden status of oldTag
		if oldTag.Hidden != t.Hidden {
			if oldTag.Hidden {
				return t.Update()
			} else {
				// oldTag.Hidden is false and t.Hidden is true
				var posts Posts
				err := posts.GetAllPostsByTag(oldTag.Id.Hex())
				if err != nil {
					return err
				}
				for _, p := range posts {
					if p.IsPublished {
						return nil
					}
				}
				return t.Update()
			}
		}
	}
	return nil
}

// GenerateTagsFromCommaString returns a slice of "Tag"s from the given input.
// The input should be a comma-seperated list of tags, like
//          "news,tech,outdoors"
func GenerateTagsFromCommaString(input string) []*Tag {
	output := make([]*Tag, 0)
	tags := strings.Split(input, ",")
	for index := range tags {
		tags[index] = strings.TrimSpace(tags[index])
	}
	for _, tag := range tags {
		if tag != "" {
			output = append(output, NewTag(tag, GenerateSlug(tag, "tags")))
		}
	}
	return output
}

// Insert inserts the tag into the DB.
func (t *Tag) Insert() error {
	// session := mdb.Copy()
	// defer session.Close()

	if len(t.Id) == 0 {
		t.Id = bson.NewObjectId()
	}
	_, err := postSession.Clone().DB(DBName).C("tags").UpsertId(t.Id, t)

	// err := meddler.Insert(db, "tags", t)
	return err
}

// Update updates an existing tag in the DB.
func (t *Tag) Update() error {
	// session := mdb.Copy()
	// defer session.Close()
	_, err := postSession.Clone().DB(DBName).C("tags").UpsertId(t.Id, t)
	// err := meddler.Insert(db, "tags", t)
	return err
}

// GetTagsByPostId finds all the tags with the give PostID
func (tags *Tags) GetTagsByPostId(postId string) error {
	session := postSession.Clone()

	// ptags := new(PostTags)
	// _ = session.DB(DBName).C("posts_tags").Find(bson.M{"postid": postId}).All(ptags)

	var ids []string
	err := session.DB(DBName).C("posts_tags").Find(bson.M{"postid": postId}).Distinct("tagid", &ids)

	// fmt.Printf("arrays %#v %v", ptags, ids)

	if err != nil {
		return err
	}

	idsbson := make([]bson.ObjectId, len(ids))
	for i, v := range ids {
		idsbson[i] = bson.ObjectIdHex(v)
	}

	// fmt.Printf("%#v\n", idsbson)

	err = session.DB(DBName).C("tags").FindId(bson.M{"$in": idsbson}).All(tags)

	// fmt.Printf("%#v %v\n", tags, err)

	// err := meddler.QueryAll(db, tags, stmtGetTagsByPostId, postId)
	return err
}

//GetTag finds any data for the tag in the DB.
func (tag *Tag) GetTag() error {
	// session := mdb.Copy()
	// defer session.Close()
	err := postSession.Clone().DB(DBName).C("tags").FindId(tag.Id).One(tag)
	// err := meddler.QueryRow(db, tag, stmtGetTag, tag.Id)
	return err
}

// GetTagBySlug finds the tag based on the Tag's slug value.
func (tag *Tag) GetTagBySlug() error {
	// session := mdb.Copy()
	// defer session.Close()
	err := postSession.Clone().DB(DBName).C("tags").Find(bson.M{"slug": tag.Slug}).One(tag)
	// err := meddler.QueryRow(db, tag, stmtGetTagBySlug, tag.Slug)
	return err
}

// GetAllTags gets all the tags in the DB.
func (tags *Tags) GetAllTags() error {
	// session := mdb.Copy()
	// defer session.Close()
	err := postSession.Clone().DB(DBName).C("tags").Find(bson.M{}).All(tags)
	// err := meddler.QueryAll(db, tags, stmtGetAllTags)
	return err
}

//DeleteOldTags removes any unused tags from the DB.
func DeleteOldTags() error {

	session := postSession.Clone()

	var ids []string
	err := session.DB(DBName).C("posts_tags").Find(bson.M{}).Distinct("tagid", &ids)
	if err != nil {
		return err
	}

	idsbson := make([]bson.ObjectId, len(ids))
	for i, v := range ids {
		idsbson[i] = bson.ObjectIdHex(v)
	}

	// fmt.Printf("%#v\n", ids)

	if len(idsbson) > 0 {
		_, err = session.DB(DBName).C("tags").RemoveAll(bson.M{"_id": bson.M{"$nin": idsbson}})
		if err == mgo.ErrNotFound {
			err = nil
		}
	} else {
		_, err = session.DB(DBName).C("tags").RemoveAll(bson.M{})
		if err == mgo.ErrNotFound {
			err = nil
		}
	}

	// WriteDB, err := db.Begin()
	// if err != nil {
	// 	WriteDB.Rollback()
	// 	return err
	// }
	// _, err = WriteDB.Exec(stmtDeleteOldTags)
	// if err != nil {
	// 	WriteDB.Rollback()
	// 	return err
	// }
	return err //WriteDB.Commit()
}

// const stmtGetTagsByPostId = `SELECT * FROM tags WHERE id IN (SELECT tag_id FROM posts_tags WHERE post_id = ?)`
// const stmtGetTag = `SELECT * FROM tags WHERE id = ?`
// const stmtGetTagBySlug = `SELECT * FROM tags WHERE slug = ?`
// const stmtGetAllTags = `SELECT * FROM tags`
// const stmtDeleteOldTags = `DELETE FROM tags WHERE id IN (SELECT id FROM tags EXCEPT SELECT tag_id FROM posts_tags)`
