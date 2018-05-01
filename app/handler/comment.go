package handler

import (
	"strconv"

	"github.com/dinever/golf"
	"github.com/covrom/dingo/app/model"
)

func registerCommentsHandlers(app *golf.Application, routes map[string]map[string]interface{}) {
	app.Get("/api/comments", APICommentsHandler)
	routes["GET"]["comments_url"] = "/api/comments"

	app.Get("/api/comments/:comment_id", APICommentHandler)
	routes["GET"]["comment_url"] = "/api/comments/:comment_id"

	app.Get("/api/comments/post/:post_id", APICommentPostHandler) // This can be removed. Use /api/posts/:post_id/comments instead.
	routes["GET"]["comment_post_url"] = "/api/comments/post/:post_id"
}

// APICommentHandler retrieves a comment with the given comment id.
func APICommentHandler(ctx *golf.Context) {
	id, err := strconv.Atoi(ctx.Param("comment_id"))
	if err != nil {
		handleErr(ctx, 500, err)
		return
	}
	comment := &model.Comment{Id: int64(id)}
	err = comment.GetCommentById()
	if err != nil {
		handleErr(ctx, 404, err)
		return
	}
	ctx.JSONIndent(comment, "", "  ")
}

// APICommentPostHandler retrives the tag with the given post id.
func APICommentPostHandler(ctx *golf.Context) {
	id, err := strconv.Atoi(ctx.Param("post_id"))
	if err != nil {
		handleErr(ctx, 500, err)
		return
	}
	comments := new(model.Comments)
	err = comments.GetCommentsByPostId(int64(id))
	if err != nil {
		handleErr(ctx, 404, err)
		return
	}
	ctx.JSONIndent(comments, "", "  ")
}

// APICommentsHandler retrieves all the comments.
func APICommentsHandler(ctx *golf.Context) {
	ctx.JSONIndent(map[string]interface{}{
		"message": "Not implemented",
	}, "", "  ")
}
