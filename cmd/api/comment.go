package main

import (
	"context"
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type commentKey string

const commentCtx commentKey = "comment"

type CommentPayload struct {
	Comment string `json:"comment"`
}

func (app *application) commentContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "commentID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()
		comment, err := app.store.Comment.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, commentCtx, comment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	idParam := chi.URLParam(r, "postID")
	postid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	user := getUserFromContext(r)

	comment := &store.Comment{
		Comment: payload.Comment,
		UserID:  user.ID,
		PostID:  postid,
	}

	ctx := r.Context()
	err = app.store.Comment.Create(ctx, comment)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	} else {
		if err := app.cacheStorage.Users.Set(ctx, 0, postid, "comment"); err != nil {
			app.logger.Error("Failed to set redis comment count Err: %v", err)
		}
	}
}

func (app *application) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	postid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()
	user := getUserFromContext(r)
	commentslice, err := app.store.Comment.GetByPostID(ctx, postid, user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, commentslice); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	commentCtx := getCommentFromCtx(r)
	if err := app.store.Comment.Update(r.Context(), commentCtx.ID, payload.Comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentCtx := getCommentFromCtx(r)
	err := app.store.Comment.Delete(r.Context(), commentCtx.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	} else {
		if err := app.cacheStorage.Users.UnSet(r.Context(), commentCtx.PostID, "", "comment"); err != nil {
			app.logger.Error("Failed to unset redis comment count %v", err)
		}
	}
}

func getCommentFromCtx(r *http.Request) *store.Comment {
	comment, _ := r.Context().Value(commentCtx).(*store.Comment)
	return comment
}
