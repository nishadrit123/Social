package main

import (
	"context"
	"database/sql"
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

// This function will no longer used, it was just created to test if user add works
func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var (
		payload CreateUserPayload
		tx      *sql.Tx
	)
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// user := getUserFromContext(r)

	user := &store.User{
		Username: payload.Username,
		// Password: payload.Password,
		// Since password is no longer a string, this is commented
		Email: payload.Email,
	}

	ctx := r.Context()
	if err := app.store.Users.Create(ctx, tx, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user, err := app.getUserFromRedisCache(r.Context(), userID)
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
	userFollowing, err := app.cacheStorage.Users.Get(r.Context(), user.ID, "", "following")
	if err != nil {
		app.logger.Error("Unable to fetch comment count from redis, Err: %v", err)
	} else {
		user.FollowingCount = userFollowing
	}
	userFollower, err := app.cacheStorage.Users.Get(r.Context(), user.ID, "", "follower")
	if err != nil {
		app.logger.Error("Unable to fetch comment count from redis, Err: %v", err)
	} else {
		user.FollowerCount = userFollower
	}
	userPostCount, err := app.cacheStorage.Users.Get(r.Context(), userID, "", "posts")
	if err != nil {
		app.logger.Error("Unable to fetch post count for user %v from redis, Err: %v", userID, err)
	} else {
		user.PostCount = userPostCount
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getUserAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	userPosts, err := app.store.Users.GetPostsByUser(r.Context(), userID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, userPosts); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, followedID, followerUser.ID); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	err = app.jsonResponse(w, http.StatusOK, nil)
	if err != nil {
		app.internalServerError(w, r, err)
	} else {
		app.cacheStorage.Users.Set(ctx, 0, followerUser.ID, "following")
		app.cacheStorage.Users.Set(ctx, 0, followedID, "follower")
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Unfollow(ctx, unfollowedID, followerUser.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusOK, nil)
	if err != nil {
		app.internalServerError(w, r, err)
	} else {
		app.cacheStorage.Users.UnSet(ctx, followerUser.ID, "", "following")
		app.cacheStorage.Users.UnSet(ctx, unfollowedID, "", "follower")
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

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, ""); err != nil {
		app.internalServerError(w, r, err)
	}
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}

func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	token := app.getJWTFromHeader(w, r)
	app.cacheStorage.Users.UnSet(r.Context(), user.ID, token, "login")
	if err := app.jsonResponse(w, http.StatusOK, ""); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getSavedPostHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	savedPostsIDS, err := app.store.Users.GetSavedPostsByUser(r.Context(), user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	var savedPosts []store.Post
	for _, savedPostsID := range savedPostsIDS {
		savedPost, err := app.store.Posts.GetByID(r.Context(), savedPostsID)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		savedPosts = append(savedPosts, *savedPost)
	}
	if err := app.jsonResponse(w, http.StatusOK, savedPosts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
