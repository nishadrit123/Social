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

		post, err := app.store.Comment.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, commentCtx, post)
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
	if err := app.store.Comment.Create(ctx, comment); err != nil {
		app.internalServerError(w, r, err)
		return
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
	commentslice, err := app.store.Comment.GetByPostID(ctx, postid)
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
	idParam := chi.URLParam(r, "commentID")
	commentid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Comment.Update(ctx, commentid, payload.Comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "commentID")
	commentID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Comment.Delete(ctx, commentID); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func getCommentFromCtx(r *http.Request) *store.Comment {
	comment, _ := r.Context().Value(commentCtx).(*store.Comment)
	return comment
}
