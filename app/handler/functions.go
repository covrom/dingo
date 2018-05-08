package handler

import (
	"github.com/covrom/dingo/app/model"
)

func getAllTags() model.Tags {
	tags := new(model.Tags)
	_ = tags.GetAllTags()
	return *tags
}

func getRecentPosts() []*model.Post {
	posts := new(model.Posts)
	_, _ = posts.GetPostList(1, 5, false, true, "published_at DESC")
	return *posts
}

func getRecentComments() []*model.Comment {
	comments := new(model.Comments)
	comments.GetCommentList(1, 5, true)
	return *comments
}
