package main

import (
	"net/http"

	"social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
		Search: "",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	user := getUserFromContext(r)
	// user := getUserFromContext(r)

	feed, err := app.store.Posts.GetUserFeed(ctx, fq, user.ID)
	for i := range feed {
		app.GetLikeCommentCountforPostwithMetadata(r, &feed[i])
	}
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
