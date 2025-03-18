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
