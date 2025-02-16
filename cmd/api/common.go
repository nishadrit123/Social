package main

import (
	"net/http"
	"social/internal/store"
)

func (app *application) GetPostsFollowersFollowingCountforUser(r *http.Request, user *store.User) {
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
	userPostCount, err := app.cacheStorage.Users.Get(r.Context(), user.ID, "", "posts")
	if err != nil {
		app.logger.Error("Unable to fetch post count for user %v from redis, Err: %v", user.ID, err)
	} else {
		user.PostCount = userPostCount
	}
}

func (app *application) GetLikeCommentCountforPost(r *http.Request, post *store.Post) {
	commentCount, err := app.cacheStorage.Users.Get(r.Context(), post.ID, "", "comment")
	if err != nil {
		app.logger.Error("Unable to fetch comment count from redis, Err: %v", err)
	} else {
		post.CommentCount = commentCount
	}
	likeCount, err := app.cacheStorage.Users.Get(r.Context(), post.ID, "", "like")
	if err != nil {
		app.logger.Error("Unable to fetch like count from redis, Err: %v", err)
	} else {
		post.LikeCount = likeCount
	}
}
