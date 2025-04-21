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

func (app *application) GetLikeCommentCountforPostwithMetadata(r *http.Request, post *store.PostWithMetadata) {
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

func (app *application) AllFollowers(w http.ResponseWriter, r *http.Request, id int64) ([]compactUserGrpPayload, error) {
	var followers []compactUserGrpPayload
	user := getUserFromContext(r)
	followerIDs, err := app.store.Users.GetFollowers(r.Context(), id)
	if err != nil {
		return followers, err
	}
	for _, followerID := range followerIDs {
		var followerInfo compactUserGrpPayload
		follower, err := app.store.Users.GetByID(r.Context(), followerID, user.ID)
		if err != nil {
			app.logger.Error("Error fetching user %v, Err: %v", followerID, err)
			continue
		}
		followerInfo.Id = follower.ID
		followerInfo.Name = follower.Username
		followers = append(followers, followerInfo)
	}
	return followers, nil
}

func (app *application) AllFollowings(w http.ResponseWriter, r *http.Request, id int64) ([]compactUserGrpPayload, error) {
	var followings []compactUserGrpPayload
	user := getUserFromContext(r)
	followingIDs, err := app.store.Users.GetFollowings(r.Context(), id)
	if err != nil {
		return followings, err
	}
	for _, followingID := range followingIDs {
		var followingInfo compactUserGrpPayload
		following, err := app.store.Users.GetByID(r.Context(), followingID, user.ID)
		if err != nil {
			app.logger.Error("Error fetching user %v, Err: %v", followingID, err)
			continue
		}
		followingInfo.Id = following.ID
		followingInfo.Name = following.Username
		followings = append(followings, followingInfo)
	}
	return followings, nil
}

func (app *application) AllGroups(w http.ResponseWriter, r *http.Request, id int64) ([]compactUserGrpPayload, error) {
	var groups []compactUserGrpPayload
	groupIDs, err := app.store.Group.GetGroupsForUser(r.Context(), id)
	if err != nil {
		return groups, nil
	}
	for _, groupID := range groupIDs {
		var groupInfo compactUserGrpPayload
		group, err := app.store.Group.GetGroupInfo(r.Context(), groupID)
		if err != nil {
			app.logger.Error("Error fetching group info for group %v, Err: %v", groupID, err)
			continue
		}
		groupInfo.Id = groupID
		groupInfo.Name = group.Name
		groupInfo.IsGroup = true
		groups = append(groups, groupInfo)
	}
	return groups, nil
}
