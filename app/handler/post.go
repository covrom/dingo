package handler

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/dinever/golf"
	"github.com/covrom/dingo/app/model"
	"github.com/covrom/dingo/app/utils"
)

func registerPostHandlers(app *golf.Application, routes map[string]map[string]interface{}) {
	adminChain := golf.NewChain(JWTAuthMiddleware)
	app.Get("/api/posts", APIPostsHandler(0, 10))
	routes["GET"]["posts_url"] = "/api/posts"

	app.Get("/api/posts/:post_id", APIPostHandler)
	routes["GET"]["post_url"] = "/api/posts/:post_id"

	app.Get("/api/posts/slug/:slug", APIPostSlugHandler)
	routes["GET"]["post_slug_url"] = "/api/posts/slug/:slug"

	app.Get("/api/posts/:post_id/comments", APIPostCommentsHandler)
	routes["GET"]["post_comments_url"] = "/api/posts/:post_id/comments"

	app.Get("/api/posts/:post_id/author", APIPostAuthorHandler)
	routes["GET"]["post_author_url"] = "/api/posts/:post_id/author"

	app.Get("/api/posts/:post_id/excerpt", APIPostExcerptHandler)
	routes["GET"]["post_excerpt_url"] = "/api/posts/:post_id/excerpt"

	app.Get("/api/posts/:post_id/summary", APIPostSummaryHandler)
	routes["GET"]["post_summary_url"] = "/api/posts/:post_id/summary"

	app.Get("/api/posts/:post_id/tag_string", APIPostTagStringHandler)
	routes["GET"]["post_tag_string_url"] = "/api/posts/:post_id/tag_string"

	app.Get("/api/posts/:post_id/tags", APIPostTagsHandler)
	routes["GET"]["post_tags_url"] = "/api/posts/:post_id/tags"

	app.Put("/api/posts", adminChain.Final(APIPostSaveHandler))
	routes["PUT"]["post_save_url"] = "/api/posts"

	app.Post("/api/posts/:post_id/publish", adminChain.Final(APIPostPublishHandler))
	routes["POST"]["post_publish_url"] = "/api/posts/:post_id/publish"

	app.Delete("/api/posts/:post_id", adminChain.Final(APIPostDeleteHandler))
	routes["DELETE"]["post_delete_url"] = "/api/posts/:post_id"
}

func getPostFromContext(ctx *golf.Context, param ...string) (post *model.Post) {
	post = new(model.Post)
	if len(param) == 0 {
		for _, p := range []string{"post_id", "slug"} {
			post = getPostFromContext(ctx, p)
			if post != nil {
				return post
			}
		}
	}
	var err error
	switch param[0] {
	case "post_id":
		id, convErr := strconv.Atoi(ctx.Param("post_id"))
		if convErr != nil {
			handleErr(ctx, 500, convErr)
			return nil
		}
		err = post.GetPostById(int64(id))
	case "slug":
		slug := ctx.Param("slug")
		err = post.GetPostBySlug(slug)
	}
	if err != nil {
		handleErr(ctx, 404, err)
		return nil
	}
	return post
}

// APIPostHandler retrieves the post with the given ID.
func APIPostHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx, "post_id")
	ctx.JSON(NewAPISuccessResponse(post))
}

// APIPostSlugHandler retrieves the post with the given slug.
func APIPostSlugHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx, "slug")
	ctx.JSON(NewAPISuccessResponse(post))
}

// APIPostsHandler gets an array of posts of length <= limit, starting at offset.
// To paginate through posts, increment offset by limit until the length of the
// post array in the response is less than limit.
func APIPostsHandler(offset, limit int) golf.HandlerFunc {
	// offset, limit args are default values
	return func(ctx *golf.Context) {
		var posts []*model.Post
		var err error
		err = ctx.Request.ParseForm()
		if err != nil {
			ctx.SendStatus(http.StatusInternalServerError)
			ctx.JSON(APIResponseBodyJSON{Status: NewErrorStatusJSON(err.Error())})
			return
		}
		if q, _ := ctx.Query("offset"); q != "" {
			offset, err = strconv.Atoi(q)
		}
		if err != nil {
			ctx.SendStatus(http.StatusBadRequest)
			ctx.JSON(APIResponseBodyJSON{Status: NewErrorStatusJSON(err.Error())})
			return
		}
		if q, _ := ctx.Query("limit"); q != "" {
			limit, err = strconv.Atoi(q)
		}
		if err != nil {
			ctx.SendStatus(http.StatusBadRequest)
			ctx.JSON(APIResponseBodyJSON{Status: NewErrorStatusJSON(err.Error())})
			return
		}
		published, _ := ctx.Query("published")
		switch published {
		case "true":
			posts, err = model.GetPublishedPosts(offset, limit)
		case "false":
			posts, err = model.GetUnpublishedPosts(offset, limit)
		default:
			posts, err = model.GetAllPosts(offset, limit)
		}
		if err != nil {
			ctx.SendStatus(http.StatusInternalServerError)
			ctx.JSON(APIResponseBodyJSON{Status: NewErrorStatusJSON(err.Error())})
			return
		}
		ctx.JSON(NewAPISuccessResponse(posts))
	}
}

// APIPostCommentsHandler gets the comments on the given post.
func APIPostCommentsHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx)
	if post == nil {
		return
	}
	comments := post.Comments()
	ctx.JSON(NewAPISuccessResponse(comments))
}

// APIPostAuthorHandler gets the author of the given post.
func APIPostAuthorHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx)
	if post == nil {
		return
	}
	author := post.Author()
	ctx.JSON(NewAPISuccessResponse(author))
}

// APIPostExcerptHandler gets the excerpt of the given post.
func APIPostExcerptHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx)
	if post == nil {
		return
	}
	excerpt := post.Excerpt()
	ctx.JSON(NewAPISuccessResponse(excerpt))
}

// APIPostSummaryHandler gets the summary of the given post.
func APIPostSummaryHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx)
	if post == nil {
		return
	}
	summary := post.Summary()
	ctx.JSON(NewAPISuccessResponse(summary))
}

// APIPostTagStringHandler gets the tag string of the given post.
func APIPostTagStringHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx)
	if post == nil {
		return
	}
	tagString := post.TagString()
	ctx.JSON(NewAPISuccessResponse(tagString))
}

// APIPostTagsHandler gets the tags of the given post.
func APIPostTagsHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx)
	if post == nil {
		return
	}
	tags := post.Tags()
	ctx.JSON(NewAPISuccessResponse(tags))
}

// APIPostSaveHandler saves the post given in the json-formatted request body.
func APIPostSaveHandler(ctx *golf.Context) {
	token, err := ctx.Session.Get("jwt")
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		return
	}
	post := model.NewPost()
	post.CreatedBy = token.(model.JWT).UserID
	defer ctx.Request.Body.Close()
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		utils.LogOnError(err, "Unable to update post from request JSON.", true)
		return
	}
	err = post.UpdateFromJSON(body)
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		ctx.JSON(APIResponseBodyJSON{Data: nil, Status: NewErrorStatusJSON(err.Error())})
		return
	}
	err = post.Save(post.Tags()...)
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		ctx.JSON(APIResponseBodyJSON{Data: nil, Status: NewErrorStatusJSON(err.Error())})
		return
	}
	ctx.JSON(NewAPISuccessResponse(post))
}

// APIPostPublishHandler publishes the post referenced by the post_id.
func APIPostPublishHandler(ctx *golf.Context) {
	token, err := ctx.Session.Get("jwt")
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		return
	}
	post := getPostFromContext(ctx)
	if post == nil {
		ctx.SendStatus(http.StatusNotFound)
		return
	}
	err = post.Publish(token.(model.JWT).UserID)
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		ctx.JSON(APIResponseBodyJSON{Data: nil, Status: NewErrorStatusJSON(err.Error())})
		return
	}
	ctx.JSON(NewAPISuccessResponse(post))
}

// APIPostDeleteHandler deletes the post referenced by the post_id.
func APIPostDeleteHandler(ctx *golf.Context) {
	post := getPostFromContext(ctx)
	if post == nil {
		ctx.SendStatus(http.StatusNotFound)
		return
	}
	err := model.DeletePostById(post.Id)
	if err != nil {
		ctx.SendStatus(http.StatusInternalServerError)
		ctx.JSON(APIResponseBodyJSON{Data: nil, Status: NewErrorStatusJSON(err.Error())})
		return
	}
}
