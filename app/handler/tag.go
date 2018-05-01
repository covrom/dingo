package handler

import (
	"strconv"

	"github.com/dinever/golf"
	"github.com/covrom/dingo/app/model"
)

func registerTagHandlers(app *golf.Application, routes map[string]map[string]interface{}) {
	app.Get("/api/tags", APITagsHandler)
	routes["GET"]["tags_url"] = "/api/tags"

	app.Get("/api/tags/:tag_id", APITagHandler)
	routes["GET"]["tag_url"] = "/api/tags/:tag_id"

	app.Get("/api/tags/slug/:slug", APITagSlugHandler)
	routes["GET"]["tag_slug_url"] = "/api/tags/:slug"
}

// APITagHandler retrieves the tag with the given id.
func APITagHandler(ctx *golf.Context) {
	id, err := strconv.Atoi(ctx.Param("tag_id"))
	if err != nil {
		handleErr(ctx, 500, err)
		return
	}
	tag := &model.Tag{Id: int64(id)}
	err = tag.GetTag()
	if err != nil {
		handleErr(ctx, 404, err)
		return
	}
	ctx.JSONIndent(tag, "", "  ")
}

// APITagsHandler retrieves all the tags.
func APITagsHandler(ctx *golf.Context) {
	tags := new(model.Tags)
	err := tags.GetAllTags()
	if err != nil {
		handleErr(ctx, 404, err)
		return
	}
	ctx.JSONIndent(tags, "", "  ")
}

// APITagSlugHandler retrieves the tag(s) with the given slug.
func APITagSlugHandler(ctx *golf.Context) {
	slug := ctx.Param("slug")
	tags := &model.Tag{Slug: slug}
	err := tags.GetTagBySlug()
	if err != nil {
		handleErr(ctx, 500, err)
		return
	}
	ctx.JSONIndent(tags, "", "  ")
}
