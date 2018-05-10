package model

import (
	"strings"

	"github.com/globalsign/mgo/bson"
)

// A Tag is a keyword associated with a post.
type Tag struct {
	Name string
	Slug string
}

// Url returns the URL of the given slug.
func (t Tag) Url() string {
	return "/tag/" + t.Slug
}

// Tags are a slice of "Tag"s
type Tags []Tag

// Len returns the amount of "Tag"s in the Tags slice.
func (t Tags) Len() int {
	return len(t)
}

// Get returns a Tag at the given index.
func (t Tags) Get(i int) Tag {
	return t[i]
}

// GetAll returns a slice of every Tag.
func (t Tags) GetAll() Tags {
	return t
}

func (t Tags) GetDistinctBySlug() (res Tags) {
	m := make(map[string]bool)
	for _, tg := range t {
		if m[tg.Slug] == true {
			continue
		}
		m[tg.Slug] = true
		res = append(res, tg)
	}
	return
}

func (t Tags) String() (res string) {
	for i, s := range t {
		res += s.Name
		if i < len(t)-1 {
			res += ", "
		}
	}
	return
}

// NewTag creates a new Tag, with CreatedAt being set to the current time.
func NewTag(name, slug string) Tag {
	return Tag{
		// Id:        bson.NewObjectId(),
		Name: name,
		Slug: slug,
		// CreatedAt: utils.Now(),
	}
}

// GenerateTagsFromCommaString returns a slice of "Tag"s from the given input.
// The input should be a comma-seperated list of tags, like
//          "news,tech,outdoors"
func GenerateTagsFromCommaString(input string) (output Tags) {
	tags := strings.Split(input, ",")
	for index := range tags {
		tags[index] = strings.TrimSpace(tags[index])
	}
	for _, tag := range tags {
		if tag != "" {
			output = append(output, NewTag(tag, GenerateSlug(tag, "tags")))
		}
	}
	return
}

// GetTagsByPostId finds all the tags with the give PostID
func (tags *Tags) GetTagsByPostId(postId string) error {

	ts := &struct{ Tags Tags }{}
	err := postSession.Clone().DB(DBName).C("posts").FindId(bson.ObjectIdHex(postId)).Select(bson.M{"tags": 1}).One(ts)
	if err == nil {
		*tags = ts.Tags.GetDistinctBySlug()
	}

	return err
}

// GetTagBySlug finds the tag based on the Tag's slug value.
func (tag *Tag) GetTagBySlug() error {
	ts := &struct{ Tags Tags }{}
	err := postSession.Clone().DB(DBName).C("posts").Find(bson.M{"tags.slug": tag.Slug}).Select(bson.M{"tags": 1}).Limit(1).One(ts)
	for _, tst := range ts.Tags {
		if tst.Slug == tag.Slug {
			*tag = tst
			break
		}
	}
	return err
}

// GetAllTags gets all the tags in the DB.
func (tags *Tags) GetAllTags() error {
	ts := &struct{ Tags Tags }{}
	err := postSession.Clone().DB(DBName).C("posts").Find(bson.M{}).Select(bson.M{"tags": 1}).One(ts)
	if err == nil {
		*tags = ts.Tags.GetDistinctBySlug()
	}
	return err
}
