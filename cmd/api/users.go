package main

import (
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

type compactUserGrpPayload struct {
	Id      int64  `json:"userid"`
	Name    string `json:"username"`
	IsGroup bool   `json:"is_group"`
}

type searchPayload struct {
	Name string `json:"name"`
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
	loggedinuser := getUserFromContext(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user, err := app.getUserFromRedisCache(r.Context(), userID, loggedinuser.ID)
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
	app.GetPostsFollowersFollowingCountforUser(r, user)

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
	user := getUserFromContext(r)
	userPosts, err := app.store.Users.GetPostsByUser(r.Context(), userID, user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	for i := range userPosts {
		app.GetLikeCommentCountforPost(r, &userPosts[i])
	}
	if err := app.jsonResponse(w, http.StatusOK, userPosts); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	var show store.DisplayButton
	followerUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if show, err = app.store.Followers.Follow(ctx, followedID, followerUser.ID); err != nil {
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
		if err := app.jsonResponse(w, http.StatusOK, show); err != nil {
			app.internalServerError(w, r, err)
		}
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	var show store.DisplayButton
	followerUser := getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if show, err = app.store.Followers.Unfollow(ctx, unfollowedID, followerUser.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.jsonResponse(w, http.StatusOK, nil)
	if err != nil {
		app.internalServerError(w, r, err)
	} else {
		app.cacheStorage.Users.UnSet(ctx, followerUser.ID, "", "following")
		app.cacheStorage.Users.UnSet(ctx, unfollowedID, "", "follower")
		if err := app.jsonResponse(w, http.StatusOK, show); err != nil {
			app.internalServerError(w, r, err)
		}
	}
}

// func (app *application) usersContextMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
// 		if err != nil {
// 			app.badRequestResponse(w, r, err)
// 			return
// 		}

// 		ctx := r.Context()

// 		// user, err := app.getUser(r.Context(), userID)
// 		user, err := app.store.Users.GetByID(r.Context(), userID)

// 		if err != nil {
// 			switch err {
// 			case store.ErrNotFound:
// 				app.notFoundResponse(w, r, err)
// 				return
// 			default:
// 				app.internalServerError(w, r, err)
// 				return
// 			}
// 		}

// 		ctx = context.WithValue(ctx, userCtx, user)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

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

func (app *application) getUserAllFollowersHandler(w http.ResponseWriter, r *http.Request) {
	var followers []compactUserGrpPayload
	idParam := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	followers, _ = app.AllFollowers(w, r, id)
	if err := app.jsonResponse(w, http.StatusOK, followers); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getUserAllFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	var followings []compactUserGrpPayload
	idParam := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	followings, _ = app.AllFollowings(w, r, id)
	if err := app.jsonResponse(w, http.StatusOK, followings); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getUserAllGroupsHandler(w http.ResponseWriter, r *http.Request) {
	var groups []compactUserGrpPayload
	idParam := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	groups, _ = app.AllGroups(w, r, id)
	if err := app.jsonResponse(w, http.StatusOK, groups); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getSavedPostHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	savedPostsIDS, err := app.store.Users.GetSavedPostsByUser(r.Context(), user.ID)
	if err != nil {
		if err := app.jsonResponse(w, http.StatusInternalServerError, err); err != nil {
			app.internalServerError(w, r, err)
		}
	}
	var savedPosts []store.Post
	for _, savedPostsID := range savedPostsIDS {
		savedPost, err := app.store.Posts.GetByID(r.Context(), savedPostsID, user.ID)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		app.GetLikeCommentCountforPost(r, savedPost)
		savedPosts = append(savedPosts, *savedPost)
	}
	if err := app.jsonResponse(w, http.StatusOK, savedPosts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) searchHandler(w http.ResponseWriter, r *http.Request) {
	var (
		payload searchPayload
	)
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	searchFor := "%" + payload.Name + "%"
	searchResult, err := app.store.Users.GetByWildCard(r.Context(), searchFor)
	if err != nil {
		app.internalServerError(w, r, err)
	}
	if err := app.jsonResponse(w, http.StatusOK, searchResult); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	token := app.getJWTFromHeader(w, r)
	app.cacheStorage.Users.UnSet(r.Context(), user.ID, token, "login")
	if err := app.jsonResponse(w, http.StatusOK, ""); err != nil {
		app.internalServerError(w, r, err)
	}
}
