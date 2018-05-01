package model

import (
	"database/sql"

	"github.com/covrom/dingo/app/utils"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const samplePostContent = `
Welcome to Dingo! This is your first post. You can find it in the [admin panel](/admin/).

Dingo uses Markdown syntax for post editing:

# Heading

## Sub-heading

### Another deeper heading

Paragraphs are separated
by a blank line.

Two spaces at the end of a line leave a
line break.

Text attributes _italic_,
**bold**, ` + "`" + `monospace` + "`" + `.

Image:

![Dingo Logo](https://cloud.githubusercontent.com/assets/1311594/14765969/bc14bafc-09c7-11e6-92f8-d69774cca249.png)

Bullet list:

  * apples
  * oranges
  * pears

Numbered list:

  1. apples
  2. oranges
  3. pears


Quoting:

> Sportsman delighted improving dashwoods gay instantly happiness six. Ham now amounted absolute not mistaken way pleasant whatever. At an these still no dried folly stood thing. Rapid it on hours hills it seven years. If polite he active county in spirit an. Mrs ham intention promotion engrossed assurance defective. Confined so graceful building opinions whatever trifling in. Insisted out differed ham man endeavor expenses. At on he total their he songs. Related compact effects is on settled do.

Code block:

` + "```" + `go
package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
` + "```" + `

Link:

An [example link](http://example.com).

Table:

|        | Cost to x | Cost to y | Cost to z |
|--------|-----------|-----------|-----------|
| From x | 0         | 3         | 4         |
| From y | 3         | 0         | 6         |
| From z | 4         | 6         | 0         |
`

// A Row contains data that can be Scanned into a variable.
type Row interface {
	Scan(dest ...interface{}) error
}

// Initialize sets up the DB by creaing a new connection, creating tables if
// they don't exist yet, and creates the welcome data.
func Initialize(dbPath string, dbExists bool) error {
	if err := initConnection(dbPath); err != nil {
		return err
	}

	if err := createTableIfNotExist(); err != nil {
		return err
	}

	if !dbExists {
		if err := createWelcomeData(); err != nil {
			return err
		}
	}

	return nil
}

func initConnection(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	return nil
}

func createTableIfNotExist() error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	checkBlogSettings()
	return nil
}

func checkBlogSettings() {
	SetSettingIfNotExists("theme", "default", "blog")
	SetSettingIfNotExists("title", "My Blog", "blog")
	SetSettingIfNotExists("description", "Awesome blog created by Dingo.", "blog")
}

func createWelcomeData() error {
	var err error
	p := NewPost()
	p.Title = "Welcome to Dingo!"
	p.Slug = "welcome-to-dingo"
	p.Markdown = samplePostContent
	p.Html = utils.Markdown2Html(p.Markdown)
	p.AllowComment = true
	p.Category = ""
	p.CreatedBy = 0
	p.UpdatedBy = 0
	p.IsPublished = true
	p.IsPage = false
	tags := GenerateTagsFromCommaString("Welcome, Dingo")
	err = p.Save(tags...)
	if err != nil {
		return err
	}

	c := NewComment()
	c.Author = "Shawn Ding"
	c.Email = "dingpeixuan911@gmail.com"
	c.Website = "http://github.com/covrom/dingo"
	c.Content = "Welcome to Dingo! This is your first comment."
	c.Avatar = utils.Gravatar(c.Email, "50")
	c.PostId = p.Id
	c.Parent = int64(0)
	c.Ip = "127.0.0.1"
	c.UserAgent = "Mozilla"
	c.UserId = 0
	c.Approved = true
	c.Save()

	SetNavigators([]string{"Home"}, []string{"/"})
	return nil
}
