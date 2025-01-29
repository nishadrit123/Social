package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"social/internal/store"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

type CreateUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Password string `json:"password" validate:"required,max=1000"`
	Email    string `json:"email"`
}

type CreateFollowerPayload struct {
	UserId int64 `json:"user_id"`
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// user := getUserFromContext(r)

	user := &store.User{
		Username: payload.Username,
		Password: payload.Password,
		Email:    payload.Email,
	}

	ctx := r.Context()
	if err := app.store.Users.Create(ctx, user); err != nil {
		fmt.Printf("err %v\n", err)
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)

	var followpayload CreateFollowerPayload
	if err := readJSON(w, r, &followpayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// if err := app.store.Followers.Follow(ctx, followedID, followerUser.ID); err != nil {
	if err := app.store.Followers.Follow(ctx, followerUser.ID, followpayload.UserId); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	// if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)

	var unfollowpayload CreateFollowerPayload
	if err := readJSON(w, r, &unfollowpayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	// if err := app.store.Followers.Unfollow(ctx, followerUser.ID, unfollowedID); err != nil {
	if err := app.store.Followers.Unfollow(ctx, followerUser.ID, unfollowpayload.UserId); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) usersContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		ctx := r.Context()

		// user, err := app.getUser(r.Context(), userID)
		user, err := app.store.Users.GetByID(r.Context(), userID)

		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFoundResponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
