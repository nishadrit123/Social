package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type storyKey string

const storyCtx storyKey = "story"

type storyPayload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (app *application) storyContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "storyID")
		story_userID := strings.Split(idParam, "-")[1]
		ctx := r.Context()

		ctx = context.WithValue(ctx, storyCtx, story_userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) createStoryHandler(w http.ResponseWriter, r *http.Request) {
	var payload storyPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	userctx := getUserFromContext(r)
	app.cacheStorage.Users.Set(r.Context(), payload, userctx.ID, "story")

	if err := app.jsonResponse(w, http.StatusCreated, payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getStoryHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	stories, err := app.cacheStorage.Users.Get(r.Context(), id, "", "story")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, stories); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deleteStoryHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "storyID")

	app.cacheStorage.Users.Delete(r.Context(), id, "story")

	if err := app.jsonResponse(w, http.StatusOK, ""); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func getStoryUserIDFromCtx(r *http.Request) int64 {
	story_UserID := r.Context().Value(storyCtx)
	id, _ := strconv.ParseInt(story_UserID.(string), 10, 64)
	return id
}
