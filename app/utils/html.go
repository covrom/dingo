package utils

import (
	"github.com/russross/blackfriday"
	"html/template"
	"regexp"
	"strings"
)

// Html2Str converts the given HTML to a string, removing all HTML tags,
// scripts, and styles, and removing continuous newline characters.
func Html2Str(html string) string {
	src := string(html)

	// Lowercase HTML tags
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)

	// Remove styles
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	// Remove scripts
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	// Strip all HTML tags
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	// Remove continuous `\n`
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")

	return strings.TrimSpace(src)
}

// SubString returns a substring of the given string from "begin" that is of
// "length" characters long.
func SubString(str string, begin, length int) (substr string) {
	rs := []rune(str)
	lth := len(rs)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}
	return string(rs[begin:end])
}

// Html2Excerpt returns the string from the given HTML that is "length" long.
func Html2Excerpt(html string, length int) string {
	return SubString(Html2Str(html), 0, length)
}

// Markdown2Html returns the given text as Markdown, using BlackFriday as a
// markdown compiler.
func Markdown2Html(text string) string {
	return string(blackfriday.Run([]byte(text)))
}

// Markdown2HtmlTemplate returns the given text as Markdown, as the template
// package's HTML type.
func Markdown2HtmlTemplate(text string) template.HTML {
	return template.HTML(Markdown2Html(text))
}
