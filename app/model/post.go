package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo"

	"net/http"

	"github.com/covrom/dingo/app/utils"
	"github.com/globalsign/mgo/bson"
)

var safeOrderByStmt = map[string]string{
	"created_at":        "createdat",
	"created_at DESC":   "-createdat",
	"updated_at":        "updatedat",
	"updated_at DESC":   "-updatedat",
	"published_at":      "publishedat",
	"published_at DESC": "-publishedat",
}

// A Post contains all the content required to populate a post or page on the
// blog. It also contains info to help sort and display the post.
type Post struct {
	Id              bson.ObjectId `bson:"_id" json:"id"`
	Title           string        `json:"title"`
	Slug            string        `json:"slug"`
	Markdown        string        `json:"markdown"`
	Html            string        `json:"html"`
	Image           string        `json:"image"`
	IsFeatured      bool          `json:"featured"`
	IsPage          bool          `json:"is_page"` // Using "is_page" instead of "page" since nouns are generally non-bools
	AllowComment    bool          `json:"allow_comment"`
	CommentNum      int64         `json:"comment_num"`
	IsPublished     bool          `json:"published"`
	Language        string        `json:"language"`
	MetaTitle       string        `json:"meta_title"`
	MetaDescription string        `json:"meta_description"`
	CreatedAt       *time.Time    `json:"created_at"`
	CreatedBy       string        `json:"created_by"`
	UpdatedAt       *time.Time    `json:"updated_at"`
	UpdatedBy       string        `json:"updated_by"`
	PublishedAt     *time.Time    `json:"published_at"`
	PublishedBy     string        `json:"published_by"`
	Tags            Tags          `json:"tags"`
	Hits            int64         `json:"-" bson:"-"`
	Category        string        `json:"-" bson:"-"`
}

// Posts is a slice of "Post"s
type Posts []*Post

// Len returns the amount of "Post"s.
func (p Posts) Len() int {
	return len(p)
}

// Get returns the Post at the given index.
func (p Posts) Get(i int) *Post {
	return p[i]
}

func (p Posts) AppendPosts(posts Posts) {
	for i := range posts {
		p = append(p, posts[i])
	}
}

// NewPost creates a new Post, with CreatedAt set to the current time.
func NewPost() *Post {
	return &Post{
		Id:        bson.NewObjectId(),
		CreatedAt: utils.Now(),
	}
}

// TagString returns all the tags associated with a post as a single string.
func (p *Post) TagString() string {

	return p.Tags.String()
}

// Url returns the URL of the post.
func (p *Post) Url() string {
	return "/" + p.Slug
}

// Author returns the User who authored the post.
func (p *Post) Author() *User {
	if !bson.IsObjectIdHex(p.CreatedBy) {
		return ghostUser
	}
	user := &User{Id: bson.ObjectIdHex(p.CreatedBy)}
	err := user.GetUserById()
	if err != nil {
		return ghostUser
	}
	return user
}

// Comments returns all the comments associated with the post.
func (p *Post) Comments() []*Comment {
	comments := new(Comments)
	err := comments.GetCommentsByPostId(p.Id.Hex())
	if err != nil {
		return nil
	}
	return comments.GetAll()
}

// Summary returns the post summary.
func (p *Post) Summary() string {
	text := strings.Split(p.Markdown, "<!--more-->")[0]
	return utils.Markdown2Html(text)
}

// Excerpt returns the post execerpt, with a default length of 255 characters.
func (p *Post) Excerpt() string {
	return utils.Html2Excerpt(p.Html, 255)
}

// Save saves a post to the DB, updating any given tags to include the Post ID.
func (p *Post) Save(tags ...Tag) error {
	p.Slug = strings.TrimLeft(p.Slug, "/")
	p.Slug = strings.TrimRight(p.Slug, "/")
	if p.Slug == "" {
		return fmt.Errorf("Slug can not be empty or root")
	}

	if p.IsPublished {
		p.PublishedAt = utils.Now()
		p.PublishedBy = p.CreatedBy
	}

	p.UpdatedAt = utils.Now()
	p.UpdatedBy = p.CreatedBy

	p.Tags = Tags(tags).GetDistinctBySlug()

	if len(p.Id) == 0 {
		// Insert post
		if err := p.Insert(); err != nil {
			return err
		}
	} else {
		if err := p.Update(); err != nil {
			return err
		}
	}
	return nil
}

// Insert saves a post to the DB.
func (p *Post) Insert() error {
	if !PostChangeSlug(p.Slug) {
		p.Slug = generateNewSlug(p.Slug, 1)
	}

	err := postSession.Clone().DB(DBName).C("posts").Insert(p)

	return err
}


// Update updates an existing post in the DB.
func (p *Post) Update() error {

	// TODO: Apply

	currentPost := &Post{Id: p.Id}
	err := currentPost.GetPostById()
	if err == mgo.ErrNotFound {
		return p.Insert()
	}
	if p.Slug != currentPost.Slug && !PostChangeSlug(p.Slug) {
		p.Slug = generateNewSlug(p.Slug, 1)
	}

	_, err = postSession.Clone().DB(DBName).C("posts").UpsertId(p.Id, p)

	return err
}

// UpdateFromRequest updates an existing Post in the DB based on the data
// provided in the HTTP request.
func (p *Post) UpdateFromRequest(r *http.Request) {
	id := r.FormValue("id")
	if bson.IsObjectIdHex(id) {
		p.Id = bson.ObjectIdHex(id)
	}
	p.Title = r.FormValue("title")
	p.Image = r.FormValue("image")
	p.Slug = r.FormValue("slug")
	p.Markdown = r.FormValue("content")
	p.Html = utils.Markdown2Html(p.Markdown)
	p.AllowComment = r.FormValue("comment") == "on"
	p.Category = r.FormValue("category")
	p.IsPublished = r.FormValue("status") == "on"
}

func (p *Post) UpdateFromJSON(j []byte) error {
	err := json.Unmarshal(j, p)
	if err != nil {
		return err
	}
	p.Html = utils.Markdown2Html(p.Markdown)
	return nil
}

func (p *Post) Publish(by string) error {
	p.PublishedAt = utils.Now()
	p.PublishedBy = by
	p.IsPublished = true

	_, err := postSession.Clone().DB(DBName).C("posts").UpsertId(p.Id, p)

	return err
}


// DeletePostById deletes the given Post from the DB.
func DeletePostById(id string) error {
	err := postSession.Clone().DB(DBName).C("posts").RemoveId(bson.ObjectIdHex(id))
	return err
}

// GetPostById gets the post based on the Post ID.
func (post *Post) GetPostById(id ...bson.ObjectId) error {
	var postId bson.ObjectId
	if len(id) == 0 {
		postId = post.Id
	} else {
		postId = id[0]
	}

	err := postSession.Clone().DB(DBName).C("posts").FindId(postId).One(post)
	return err
}

// GetPostBySlug gets the post based on the Post Slug.
func (p *Post) GetPostBySlug(slug string) error {
	err := postSession.Clone().DB(DBName).C("posts").Find(bson.M{"slug": slug}).One(p)
	return err
}

// GetPostsByTag returns a new pager based all the Posts associated with a Tag.
func (p *Posts) GetPostsByTag(tagslug string, page, size int64, onlyPublished bool) (*utils.Pager, error) {
	var (
		pager *utils.Pager
		count int64
	)

	session := postSession.Clone()

	cnt, err := session.DB(DBName).C("posts").Find(bson.M{"tags.slug": tagslug}).Count()
	if err != nil {
		utils.LogOnError(err, "Unable to get posts by tag.", true)
		return nil, err
	}
	count = int64(cnt)

	pager = utils.NewPager(page, size, count)

	if !pager.IsValid {
		return pager, fmt.Errorf("Page not found")
	}

	if onlyPublished {
		err = session.DB(DBName).C("posts").Find(bson.M{"tags.slug": tagslug, "ispublished": true}).Sort("-publishedat").Skip(int(pager.Begin)).Limit(int(size)).All(p)
	} else {
		err = session.DB(DBName).C("posts").Find(bson.M{"tags.slug": tagslug}).Sort("-publishedat").Skip(int(pager.Begin)).Limit(int(size)).All(p)
	}
	return pager, err
}

// GetAllPostsByTag gets all the Posts with the associated Tag.
func (p *Posts) GetAllPostsByTag(tag Tag) error {
	session := postSession.Clone()

	err := session.DB(DBName).C("posts").Find(bson.M{"tags": tag}).Sort("-publishedat").All(p)

	return err
}

// GetNumberOfPosts gets the total number of posts in the DB.
func GetNumberOfPosts(isPage bool, published bool) (int64, error) {
	session := postSession.Clone()

	var err error
	var cnt int
	if published {
		cnt, err = session.DB(DBName).C("posts").Find(bson.M{"ispage": isPage, "ispublished": true}).Count()
	} else {
		cnt, err = session.DB(DBName).C("posts").Find(bson.M{"ispage": isPage}).Count()
	}

	return int64(cnt), err
}

// GetPostList returns a new pager based on all the posts in the DB.
func (posts *Posts) GetPostList(page, size int64, isPage bool, onlyPublished bool, orderBy string) (*utils.Pager, error) {
	var pager *utils.Pager
	count, err := GetNumberOfPosts(isPage, onlyPublished)
	pager = utils.NewPager(page, size, count)

	if !pager.IsValid {
		return pager, fmt.Errorf("Page not found")
	}

	safeOrderBy := getSafeOrderByStmt(orderBy)

	session := postSession.Clone()

	if onlyPublished {
		err = session.DB(DBName).C("posts").Find(bson.M{"ispage": isPage, "ispublished": true}).Sort(safeOrderBy).Skip(int(pager.Begin)).Limit(int(size)).All(posts)
	} else {
		err = session.DB(DBName).C("posts").Find(bson.M{"ispage": isPage}).Sort(safeOrderBy).Skip(int(pager.Begin)).Limit(int(size)).All(posts)
	}

	return pager, err
}

// GetAllPostList gets all the posts, with the options to get only pages, or
// only published posts. It is also possible to order the posts, with the order
// by string being one of six options:
//         "created_at"
//         "created_at DESC"
//         "updated_at"
//         "updated_at DESC"
//         "published_at"
//         "published_at DESC"
func (p *Posts) GetAllPostList(isPage bool, onlyPublished bool, orderBy string) error {
	session := postSession.Clone()

	var err error

	safeOrderBy := getSafeOrderByStmt(orderBy)

	if onlyPublished {
		err = session.DB(DBName).C("posts").Find(bson.M{"ispage": isPage, "ispublished": true}).Sort(safeOrderBy).All(p)
	} else {
		err = session.DB(DBName).C("posts").Find(bson.M{"ispage": isPage}).Sort(safeOrderBy).All(p)
	}

	return err
}

// PostChangeSlug checks to see if there is a post associated with the given
// slug, and returns true if there isn't.
func PostChangeSlug(slug string) bool {
	post := new(Post)
	err := post.GetPostBySlug(slug)
	if err != nil {
		return true
	}
	return false
}

func generateNewSlug(slug string, suffix int) string {
	newSlug := slug + "-" + strconv.Itoa(suffix)
	if !PostChangeSlug(newSlug) {
		return generateNewSlug(slug, suffix+1)
	}
	return newSlug
}

// getSafeOrderByStmt returns a safe `ORDER BY` statement to be when used when
// building SQL queries, in order to prevent SQL injection.
//
// Since we can't use the placeholder `?` to specify the `ORDER BY` values in
// queries, we need to build them using `fmt.Sprintf`. Typically, doing so
// would open you up to SQL injection attacks, since any string can be passed
// into `fmt.Sprintf`, including strings that are valid SQL queries! By using
// this function to check a map of safe values, we guarantee that no unsafe
// values are ever passed to our query building function.
func getSafeOrderByStmt(orderBy string) string {
	if stmt, ok := safeOrderByStmt[orderBy]; ok {
		return stmt
	}
	return "-publishedat"
}

func GetPublishedPosts(offset, limit int) (Posts, error) {

	var posts Posts
	err := postSession.Clone().DB(DBName).C("posts").Find(bson.M{"ispublished": true}).Skip(offset).Limit(limit).All(&posts)

	return posts, err
}

func GetUnpublishedPosts(offset, limit int) (Posts, error) {
	var posts Posts
	err := postSession.Clone().DB(DBName).C("posts").Find(bson.M{"ispublished": false}).Skip(offset).Limit(limit).All(&posts)
	return posts, err
}

func GetAllPosts(offset, limit int) ([]*Post, error) {
	pubPosts, err := GetPublishedPosts(offset, limit)
	if err != nil {
		return nil, err
	}
	unpubPosts, err := GetUnpublishedPosts(offset, limit)
	if err != nil {
		return nil, err
	}
	posts := append(pubPosts, unpubPosts...)
	return posts, nil
}
