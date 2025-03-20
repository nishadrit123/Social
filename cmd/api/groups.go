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

func (app *application) getGroupMembers(w http.ResponseWriter, r *http.Request) {
	var (
		member      *store.User
		memberSlice []store.User
		userSlice   []int64
	)
	idParam := chi.URLParam(r, "groupID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	userSlice, err = app.store.Group.GetGroupMembers(r.Context(), id)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	for _, userID := range userSlice {
		member, err = app.store.Users.GetByID(r.Context(), userID)
		if err != nil {
			app.logger.Errorf("Error while fetching member %v from group %v, Err: %v", userID, id, err)
			continue
		}
		memberSlice = append(memberSlice, *member)
	}

	if err := app.jsonResponse(w, http.StatusOK, memberSlice); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
