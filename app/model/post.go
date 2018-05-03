package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/covrom/dingo/app/utils"
	"github.com/globalsign/mgo/bson"
	// "github.com/russross/meddler"
)

const stmtGetPostById = `SELECT * FROM posts WHERE id = ?`
const stmtGetPostBySlug = `SELECT * FROM posts WHERE slug = ?`
const stmtGetPostsByTag = `SELECT * FROM posts WHERE %s id IN ( SELECT post_id FROM posts_tags WHERE tag_id = ? ) ORDER BY published_at DESC LIMIT ? OFFSET ?`
const stmtGetAllPostsByTag = `SELECT * FROM posts WHERE id IN ( SELECT post_id FROM posts_tags WHERE tag_id = ?) ORDER BY published_at DESC `
const stmtGetPostsCountByTag = "SELECT count(*) FROM posts, posts_tags WHERE posts_tags.post_id = posts.id AND posts.published AND posts_tags.tag_id = ?"
const stmtGetPostsOffsetLimit = `SELECT * FROM posts WHERE published = ? LIMIT ?, ?`
const stmtInsertPostTag = `INSERT INTO posts_tags (id, post_id, tag_id) VALUES (?, ?, ?)`
const stmtDeletePostTagsByPostId = `DELETE FROM posts_tags WHERE post_id = ?`
const stmtNumberOfPosts = "SELECT count(*) FROM posts WHERE %s"
const stmtGetAllPostList = `SELECT * FROM posts WHERE %s ORDER BY %s`
const stmtGetPostList = `SELECT * FROM posts WHERE %s ORDER BY %s LIMIT ? OFFSET ?`
const stmtDeletePostById = `DELETE FROM posts WHERE id = ?`

var safeOrderByStmt = map[string]string{
	"created_at":        "created_at",
	"created_at DESC":   "-created_at",
	"updated_at":        "updated_at",
	"updated_at DESC":   "-updated_at",
	"published_at":      "published_at",
	"published_at DESC": "-published_at",
}

// A Post contains all the content required to populate a post or page on the
// blog. It also contains info to help sort and display the post.
type Post struct {
	Id              int64      `json:"id"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Markdown        string     `json:"markdown"`
	Html            string     `json:"html"`
	Image           string     `json:"image"`
	IsFeatured      bool       `json:"featured"`
	IsPage          bool       `json:"is_page"` // Using "is_page" instead of "page" since nouns are generally non-bools
	AllowComment    bool       `json:"allow_comment"`
	CommentNum      int64      `json:"comment_num"`
	IsPublished     bool       `json:"published"`
	Language        string     `json:"language"`
	MetaTitle       string     `json:"meta_title"`
	MetaDescription string     `json:"meta_description"`
	CreatedAt       *time.Time `json:"created_at"`
	CreatedBy       int64      `json:"created_by"`
	UpdatedAt       *time.Time `json:"updated_at"`
	UpdatedBy       int64      `json:"updated_by"`
	PublishedAt     *time.Time `json:"published_at"`
	PublishedBy     int64      `json:"published_by"`
	Hits            int64      `json:"-"`
	Category        string     `json:"-"`
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
		CreatedAt: utils.Now(),
	}
}

// TagString returns all the tags associated with a post as a single string.
func (p *Post) TagString() string {
	tags := new(Tags)
	_ = tags.GetTagsByPostId(p.Id)
	var tagString string
	for i := 0; i < tags.Len(); i++ {
		if i != tags.Len()-1 {
			tagString += tags.Get(i).Name + ", "
		} else {
			tagString += tags.Get(i).Name
		}
	}
	return tagString
}

// Url returns the URL of the post.
func (p *Post) Url() string {
	return "/" + p.Slug
}

// Tags returns a slice of every tag associated with the post.
func (p *Post) Tags() []*Tag {
	tags := new(Tags)
	err := tags.GetTagsByPostId(p.Id)
	if err != nil {
		return nil
	}
	return tags.GetAll()
}

// Author returns the User who authored the post.
func (p *Post) Author() *User {
	user := &User{Id: p.CreatedBy}
	err := user.GetUserById()
	if err != nil {
		return ghostUser
	}
	return user
}

// Comments returns all the comments associated with the post.
func (p *Post) Comments() []*Comment {
	comments := new(Comments)
	err := comments.GetCommentsByPostId(p.Id)
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
func (p *Post) Save(tags ...*Tag) error {
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

	if p.Id == 0 {
		// Insert post
		if err := p.Insert(); err != nil {
			return err
		}
	} else {
		if err := p.Update(); err != nil {
			return err
		}
	}
	tagIds := make([]int64, 0)
	// Insert tags
	for _, t := range tags {
		t.CreatedAt = utils.Now()
		t.CreatedBy = p.CreatedBy
		t.Hidden = !p.IsPublished
		t.Save()
		tagIds = append(tagIds, t.Id)
	}
	// Delete old post-tag projections
	err := DeletePostTagsByPostId(p.Id)
	// Insert postTags
	if err != nil {
		return err
	}
	for _, tagId := range tagIds {
		err := InsertPostTag(p.Id, tagId)
		if err != nil {
			return err
		}
	}
	return DeleteOldTags()
}

// Insert saves a post to the DB.
func (p *Post) Insert() error {
	if !PostChangeSlug(p.Slug) {
		p.Slug = generateNewSlug(p.Slug, 1)
	}
	session := mdb.Copy()
	defer session.Close()

	err := session.DB(DBName).C("posts").Insert(p)

	// err := meddler.Insert(db, "posts", p)
	return err
}

type PostTag struct {
	PostId int64 `json:"post_id"`
	TagId  int64 `json:"tag_id"`
}

// InsertPostTag saves the Post ID to the given Tag ID in the DB.
func InsertPostTag(postID int64, tagID int64) error {
	// writeDB, err := db.Begin()
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	// _, err = writeDB.Exec(stmtInsertPostTag, nil, postID, tagID)
	session := mdb.Copy()
	defer session.Close()

	err := session.DB(DBName).C("posts_tags").Insert(&PostTag{PostId: postID, TagId: tagID})
	if err != nil {
		// writeDB.Rollback()
		return err
	}
	return nil //writeDB.Commit()
}

// Update updates an existing post in the DB.
func (p *Post) Update() error {

	// TODO: Apply

	currentPost := &Post{Id: p.Id}
	err := currentPost.GetPostById()
	if err != nil {
		return err
	}
	if p.Slug != currentPost.Slug && !PostChangeSlug(p.Slug) {
		p.Slug = generateNewSlug(p.Slug, 1)
	}

	session := mdb.Copy()
	defer session.Close()
	_, err = session.DB(DBName).C("posts").Upsert(bson.M{"id": p.Id}, p)

	// err = meddler.Update(db, "posts", p)
	return err
}

// UpdateFromRequest updates an existing Post in the DB based on the data
// provided in the HTTP request.
func (p *Post) UpdateFromRequest(r *http.Request) {
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

func (p *Post) Publish(by int64) error {
	p.PublishedAt = utils.Now()
	p.PublishedBy = by
	p.IsPublished = true

	session := mdb.Copy()
	defer session.Close()
	_, err := session.DB(DBName).C("posts").Upsert(bson.M{"id": p.Id}, p)

	// err := meddler.Update(db, "posts", p)
	return err
}

// DeletePostTagsByPostId deletes removes tags associated with the given post
// from the DB.
func DeletePostTagsByPostId(post_id int64) error {
	session := mdb.Copy()
	defer session.Close()
	err := session.DB(DBName).C("posts_tags").Remove(bson.M{"post_id": post_id})

	// writeDB, err := db.Begin()
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	// _, err = writeDB.Exec(stmtDeletePostTagsByPostId, post_id)
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	return err //writeDB.Commit()
}

// DeletePostById deletes the given Post from the DB.
func DeletePostById(id int64) error {
	session := mdb.Copy()
	defer session.Close()
	err := session.DB(DBName).C("posts").Remove(bson.M{"id": id})

	// writeDB, err := db.Begin()
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	// _, err = writeDB.Exec(stmtDeletePostById, id)
	// if err != nil {
	// 	writeDB.Rollback()
	// 	return err
	// }
	// err = writeDB.Commit()

	if err != nil {
		return err
	}
	err = DeletePostTagsByPostId(id)
	if err != nil {
		return err
	}
	return DeleteOldTags()
}

// GetPostById gets the post based on the Post ID.
func (post *Post) GetPostById(id ...int64) error {
	var postId int64
	if len(id) == 0 {
		postId = post.Id
	} else {
		postId = id[0]
	}
	session := mdb.Copy()
	defer session.Close()

	err := session.DB(DBName).C("posts").Find(bson.M{"id": postId}).One(post)
	// err := meddler.QueryRow(db, post, stmtGetPostById, postId)
	return err
}

// GetPostBySlug gets the post based on the Post Slug.
func (p *Post) GetPostBySlug(slug string) error {
	session := mdb.Copy()
	defer session.Close()
	err := session.DB(DBName).C("posts").Find(bson.M{"slug": slug}).One(p)

	// err := meddler.QueryRow(db, p, stmtGetPostBySlug, slug)
	return err
}

type PostTags []*PostTag

// GetPostsByTag returns a new pager based all the Posts associated with a Tag.
func (p *Posts) GetPostsByTag(tagId, page, size int64, onlyPublished bool) (*utils.Pager, error) {
	var (
		pager *utils.Pager
		count int64
	)

	session := mdb.Copy()
	defer session.Close()

	ptags := new(PostTags)
	err := session.DB(DBName).C("posts_tags").Find(bson.M{"tag_id": tagId}).All(ptags)
	if err != nil {
		utils.LogOnError(err, "Unable to get posts by tag.", true)
		return nil, err
	}

	ids := make([]int64, len(*ptags))
	if len(*ptags) > 0 {
		for i, v := range *ptags {
			ids[i] = v.PostId
		}
		cnt, err := session.DB(DBName).C("posts").Find(bson.M{"id": bson.M{"$in": ids}}).Count()
		if err != nil {
			utils.LogOnError(err, "Unable to get posts by tag.", true)
			return nil, err
		}
		count = int64(cnt)
	}

	// row := db.QueryRow(stmtGetPostsCountByTag, tagId)
	// err := row.Scan(&count)
	// if err != nil {
	// 	utils.LogOnError(err, "Unable to get posts by tag.", true)
	// 	return nil, err
	// }

	pager = utils.NewPager(page, size, count)

	if !pager.IsValid {
		return pager, fmt.Errorf("Page not found")
	}

	// var where string
	if onlyPublished {
		// where = "published AND"
		err = session.DB(DBName).C("posts").Find(bson.M{"id": bson.M{"$in": ids}, "published": true}).Sort("-published_at").Skip(int(pager.Begin)).Limit(int(size)).All(p)
	} else {
		err = session.DB(DBName).C("posts").Find(bson.M{"id": bson.M{"$in": ids}}).Sort("-published_at").Skip(int(pager.Begin)).Limit(int(size)).All(p)
	}
	// err = meddler.QueryAll(db, p, fmt.Sprintf(stmtGetPostsByTag, where), tagId, size, pager.Begin)
	return pager, err
}

// GetAllPostsByTag gets all the Posts with the associated Tag.
func (p *Posts) GetAllPostsByTag(tagId int64) error {
	session := mdb.Copy()
	defer session.Close()

	ptags := new(PostTags)
	err := session.DB(DBName).C("posts_tags").Find(bson.M{"tag_id": tagId}).All(ptags)
	if err != nil {
		return err
	}

	ids := make([]int64, len(*ptags))
	for i, v := range *ptags {
		ids[i] = v.PostId
	}
	err = session.DB(DBName).C("posts").Find(bson.M{"id": bson.M{"$in": ids}}).Sort("-published_at").All(p)

	return err
}

// GetNumberOfPosts gets the total number of posts in the DB.
func GetNumberOfPosts(isPage bool, published bool) (int64, error) {
	session := mdb.Copy()
	defer session.Close()

	var err error
	var cnt int
	if published {
		cnt, err = session.DB(DBName).C("posts").Find(bson.M{"is_page": isPage, "published": true}).Count()
	} else {
		cnt, err = session.DB(DBName).C("posts").Find(bson.M{"is_page": isPage}).Count()
	}

	// var count int64
	// var where string
	// if isPage {
	// 	where = `page = 1`
	// } else {
	// 	where = `page = 0`
	// }
	// if published {
	// 	where = where + ` AND published`
	// }
	// var row *sql.Row

	// row = db.QueryRow(fmt.Sprintf(stmtNumberOfPosts, where))
	// err := row.Scan(&count)
	// if err != nil {
	// 	return 0, err
	// }
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

	session := mdb.Copy()
	defer session.Close()

	if onlyPublished {
		err = session.DB(DBName).C("posts").Find(bson.M{"is_page": isPage, "published": true}).Sort(safeOrderBy).Skip(int(pager.Begin)).Limit(int(size)).All(posts)
	} else {
		err = session.DB(DBName).C("posts").Find(bson.M{"is_page": isPage}).Sort(safeOrderBy).Skip(int(pager.Begin)).Limit(int(size)).All(posts)
	}

	// var where string
	// if isPage {
	// 	where = `page = 1`
	// } else {
	// 	where = `page = 0`
	// }
	// if onlyPublished {
	// 	where = where + ` AND published`
	// }
	// err = meddler.QueryAll(db, posts, fmt.Sprintf(stmtGetPostList, where, safeOrderBy), size, pager.Begin)
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
	// var where string
	// if isPage {
	// 	where = `page = 1`
	// } else {
	// 	where = `page = 0`
	// }
	// if onlyPublished {
	// 	where = where + ` AND published`
	// }

	session := mdb.Copy()
	defer session.Close()

	var err error

	safeOrderBy := getSafeOrderByStmt(orderBy)

	if onlyPublished {
		err = session.DB(DBName).C("posts").Find(bson.M{"is_page": isPage, "published": true}).Sort(safeOrderBy).All(p)
	} else {
		err = session.DB(DBName).C("posts").Find(bson.M{"is_page": isPage}).Sort(safeOrderBy).All(p)
	}

	// err := meddler.QueryAll(db, p, fmt.Sprintf(stmtGetAllPostList, where, safeOrderBy))
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
	return "-published_at"
}

func GetPublishedPosts(offset, limit int) (Posts, error) {
	session := mdb.Copy()
	defer session.Close()

	var posts Posts
	err := session.DB(DBName).C("posts").Find(bson.M{"published": true}).Skip(offset).Limit(limit).All(&posts)

	// err := meddler.QueryAll(db, &posts, stmtGetPostsOffsetLimit, 1, offset, limit)
	return posts, err
}

func GetUnpublishedPosts(offset, limit int) (Posts, error) {
	session := mdb.Copy()
	defer session.Close()

	var posts Posts
	err := session.DB(DBName).C("posts").Find(bson.M{"published": false}).Skip(offset).Limit(limit).All(&posts)
	// err := meddler.QueryAll(db, &posts, stmtGetPostsOffsetLimit, 0, offset, limit)
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
