package main

import (
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type groupMembersPayload struct {
	Name    string  `json:"name,omitempty"`
	Members []int64 `json:"members"`
}

type groupInfoPayload struct {
	Name    string       `json:"name,omitempty"`
	Admin   *store.User  `json:"admin"`
	Members []store.User `json:"members"`
}

func (app *application) createGroupHandler(w http.ResponseWriter, r *http.Request) {
	var payload groupMembersPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	grp := &store.Group{
		Name:      payload.Name,
		CreatedBy: user.ID,
		Members:   payload.Members,
	}

	ctx := r.Context()
	if err := app.store.Group.Create(ctx, grp); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, grp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) addGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	var payload groupMembersPayload

	idParam := chi.URLParam(r, "groupID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	grp := &store.Group{
		Members: payload.Members,
	}

	ctx := r.Context()
	if err := app.store.Group.AddMembers(ctx, id, grp); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, "Added"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getGroupInfo(w http.ResponseWriter, r *http.Request) {
	var (
		member    *store.User
		groupInfo groupInfoPayload
	)
	idParam := chi.URLParam(r, "groupID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	user := getUserFromContext(r)
	isMember, err := app.store.Group.IsUserInGroup(r.Context(), id, user.ID)
	if !isMember || err != nil {
		if err := app.jsonResponse(w, http.StatusUnauthorized, err); err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}
		return
	}

	grpInfo, err := app.store.Group.GetGroupInfo(r.Context(), id)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	groupInfo.Name = grpInfo.Name
	admin, err := app.store.Users.GetByID(r.Context(), grpInfo.CreatedBy)
	if err != nil {
		app.logger.Errorf("Error while fetching admin %v from group %v, Err: %v", grpInfo.CreatedBy, id, err)
	}
	groupInfo.Admin = admin
	for _, userID := range grpInfo.Members {
		member, err = app.store.Users.GetByID(r.Context(), userID)
		if err != nil {
			app.logger.Errorf("Error while fetching member %v from group %v, Err: %v", userID, id, err)
			continue
		}
		groupInfo.Members = append(groupInfo.Members, *member)
	}

	if err := app.jsonResponse(w, http.StatusOK, groupInfo); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
