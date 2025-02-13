package main

import (
	"net/http"
)

func (app *application) likedislikeHandler(w http.ResponseWriter, r *http.Request) {
	postctx := getPostFromCtx(r)
	userCtx := getUserFromContext(r)
	exists, err := app.store.Like.LikeDislike(r.Context(), postctx.ID, userCtx.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	} else {
		if !exists {
			app.cacheStorage.Users.Set(r.Context(), 0, postctx.ID, "like")
		} else {
			app.cacheStorage.Users.UnSet(r.Context(), postctx.ID, "", "like")
		}
	}
}
