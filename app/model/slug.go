package model

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/globalsign/mgo"
)

// GenerateSlug generates a URL-friendly slug. The table is one of "posts",
// "tags", "navigation", or "users".
func GenerateSlug(input string, table string) string {
	output := strings.Map(func(r rune) rune {
		switch {
		case r == ' ', r == '-':
			return '-'
		case r == '_', unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		default:
			return -1
		}
	}, strings.ToLower(strings.TrimSpace(input)))
	// Maximum of 75 characters for slugs right now
	maxLength := 75
	if len([]rune(output)) > maxLength {
		runes := []rune(output)[:maxLength]
		// Try to cut at '-' until length of (maxLength - (maxLength / 2)) characters
		for i := (maxLength - 1); i > (maxLength - (maxLength / 2)); i-- {
			if runes[i] == '-' {
				runes = runes[:i]
				break
			}
		}
		output = string(runes)
	}
	// Don't allow a few specific slugs that are used by the blog
	if table == "posts" && (output == "rss" || output == "tag" || output == "author" || output == "page" || output == "admin") {
		output = generateUniqueSlug(output, table, 2)
	} else if table == "tags" || table == "navigation" { // We want duplicate tag and navigation slugs
		return output
	}
	return generateUniqueSlug(output, table, 1)
}

func generateUniqueSlug(slug string, table string, suffix int) string {
	// Recursive function
	slugToCheck := slug
	if suffix > 1 { // If this is not the first try, add the suffix and try again
		slugToCheck = slug + "-" + strconv.Itoa(suffix)
	}
	var err error
	if table == "tags" { // Not needed at the moment. Tags with the same name should have the same slug.
		// tag := &Tag{Slug: slugToCheck}
		err = mgo.ErrNotFound //tag.GetTagBySlug()
	} else if table == "posts" {
		post := new(Post)
		err = post.GetPostBySlug(slugToCheck)
	} else if table == "users" {
		u := new(User)
		err = u.GetUserBySlug()
	}
	if err == nil {
		return generateUniqueSlug(slug, table, suffix+1)
	}
	return slugToCheck
}
