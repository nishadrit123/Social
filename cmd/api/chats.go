package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type chatPayload struct {
	SenderID   int64     `json:"sender_id,omitempty"`
	ReceiverID int64     `json:"receiver_id,omitempty"` // can be both userId or groupID
	Text       string    `json:"text,omitempty"`
	PostID     int64     `json:"post_id,omitempty"`
	Date       time.Time `json:"date,omitempty"`
}

func (app *application) postChat(w http.ResponseWriter, r *http.Request) {
	var (
		subject string
		idParam string
		is_user bool
	)
	if strings.Contains(r.URL.String(), "user") {
		is_user = true
	}
	if is_user {
		idParam = chi.URLParam(r, "userID")
	} else {
		idParam = chi.URLParam(r, "groupID")
	}
	receiverID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	sender := getUserFromContext(r)
	senderID := sender.ID

	var payload chatPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	payload.SenderID = senderID
	payload.ReceiverID = receiverID
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		app.logger.Errorf("Error loading IST time, Err: %v", err)
	}
	payload.Date = time.Now().In(loc)

	bytePayload, err := json.Marshal(payload)
	if err != nil {
		app.logger.Errorf("Error marshalling chat payload, Err: %v", err)
	}

	if is_user {
		if senderID < receiverID {
			subject = fmt.Sprintf("chat.user.%v.%v", senderID, receiverID)
		} else {
			subject = fmt.Sprintf("chat.user.%v.%v", receiverID, senderID)
		}
	} else {
		isMember, err := app.store.Group.IsUserInGroup(r.Context(), receiverID, senderID)
		if !isMember || err != nil {
			if err := app.jsonResponse(w, http.StatusUnauthorized, err); err != nil {
				app.unauthorizedErrorResponse(w, r, err)
				return
			}
			return
		}
		subject = fmt.Sprintf("chat.group.%v", receiverID)
	}
	err = app.nats.NatsConn.SendToChat(subject, bytePayload, is_user)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// since a text message is being sent, all the previous chats/posts are supposed to be displayed unlike in
	// case of sending a post where no previous chats/posts are supposed to be displayed
	if payload.Text != "" {
		app.FetchAllChats(w, r, subject, is_user)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, "SENT"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getChat(w http.ResponseWriter, r *http.Request) {
	var (
		subject string
		idParam string
		is_user bool
	)
	if strings.Contains(r.URL.String(), "user") {
		is_user = true
	}
	if is_user {
		idParam = chi.URLParam(r, "userID")
	} else {
		idParam = chi.URLParam(r, "groupID")
	}
	otherID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	loggedInUser := getUserFromContext(r)
	loggedInUserID := loggedInUser.ID

	if is_user {
		if loggedInUserID < otherID {
			subject = fmt.Sprintf("chat.user.%v.%v", loggedInUserID, otherID)
		} else {
			subject = fmt.Sprintf("chat.user.%v.%v", otherID, loggedInUserID)
		}
	} else {
		isMember, err := app.store.Group.IsUserInGroup(r.Context(), otherID, loggedInUserID)
		if !isMember || err != nil {
			if err := app.jsonResponse(w, http.StatusUnauthorized, err); err != nil {
				app.unauthorizedErrorResponse(w, r, err)
				return
			}
			return
		}
		subject = fmt.Sprintf("chat.group.%v", otherID)
	}

	app.FetchAllChats(w, r, subject, is_user)
}

func (app *application) FetchAllChats(w http.ResponseWriter, r *http.Request, subject string, is_user bool) {
	allChats, err := app.nats.NatsConn.GetallChats(subject, is_user)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	for i := range allChats {
		if allChats[i].PostID != 0 {
			post, err := app.store.Posts.GetByID(r.Context(), allChats[i].PostID)
			if err != nil {
				app.logger.Errorf("Error fetching post by id for chats, Err: %v", err)
			}
			app.GetLikeCommentCountforPost(r, post)
			bytePost, err := json.Marshal(post)
			if err != nil {
				app.logger.Errorf("Error marshalling post for chat, Err: %v", err)
			}
			err = json.Unmarshal(bytePost, &allChats[i].ChatPost)
			if err != nil {
				app.logger.Errorf("Error unmarshalling post for chat, Err: %v", err)
			}
		}
	}
	if err := app.jsonResponse(w, http.StatusOK, allChats); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
