package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"social/internal/store"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  user.ID,
		// UserID: payload.Userid,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	app.cacheStorage.Users.Set(ctx, 0, user.ID, "posts")

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	ctxpost := getPostFromCtx(r)
	user := getUserFromContext(r)
	post, err := app.store.Posts.GetByID(r.Context(), ctxpost.ID, user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.GetLikeCommentCountforPost(r, post)

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getUserofPostHandler(w http.ResponseWriter, r *http.Request) {
	ctxpost := getPostFromCtx(r)
	user := getUserFromContext(r)
	user, err := app.store.Users.GetByID(r.Context(), ctxpost.UserID, user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	app.GetPostsFollowersFollowingCountforUser(r, user)
	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) createSaveUnsavePostHandler(w http.ResponseWriter, r *http.Request) {
	ctxuser := getUserFromContext(r)
	ctxpost := getPostFromCtx(r)
	_, err := app.store.Users.SaveUnsavePost(r.Context(), ctxuser.ID, ctxpost.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, err); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getSendPostHandler(w http.ResponseWriter, r *http.Request) {
	var sendTo []compactUserGrpPayload
	ctxuser := getUserFromContext(r)

	followers, _ := app.AllFollowers(w, r, ctxuser.ID)
	followings, _ := app.AllFollowings(w, r, ctxuser.ID)
	groups, _ := app.AllGroups(w, r, ctxuser.ID)

	unique_friends := make(map[compactUserGrpPayload]bool)
	for _, follower_friend := range followers {
		unique_friends[follower_friend] = true
	}
	for _, following_fried := range followings {
		if _, exists := unique_friends[following_fried]; !exists {
			unique_friends[following_fried] = true
		}
	}

	// if len(followers) > 0 {
	// 	sendTo = append(sendTo, followers...)
	// }
	// if len(followings) > 0 {
	// 	sendTo = append(sendTo, followings...)
	// }
	// followers = append(followers, followings...)
	// for _, all := range followers {
	// 	if _, exists := unique_friends[all.Name]; exists {
	// 		sendTo = append(sendTo, all)
	// 		unique_friends[all.Name] = false
	// 	}
	// }
	for friend := range unique_friends {
		sendTo = append(sendTo, friend)
	}
	if len(groups) > 0 {
		sendTo = append(sendTo, groups...)
	}

	if err := app.jsonResponse(w, http.StatusOK, sendTo); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	user := getUserFromContext(r)

	if err := app.store.Posts.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.store.Comment.DeleteByPostID(ctx, id)
	app.store.Like.DeleteByPostID(ctx, id)

	app.cacheStorage.Users.Delete(ctx, id, "comment")
	app.cacheStorage.Users.Delete(ctx, id, "like")
	app.cacheStorage.Users.UnSet(ctx, user.ID, "", "posts")

	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string  `json:"title" validate:"omitempty,max=100"`
	Content *string  `json:"content" validate:"omitempty,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}

	ctx := r.Context()

	if err := app.updatePost(ctx, post); err != nil {
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()
		user := getUserFromContext(r)

		post, err := app.store.Posts.GetByID(ctx, id, user.ID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}

func (app *application) updatePost(ctx context.Context, post *store.Post) error {
	if err := app.store.Posts.Update(ctx, post); err != nil {
		return err
	}
	return nil
}
