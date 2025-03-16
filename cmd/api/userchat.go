package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func (app *application) postUserChat(w http.ResponseWriter, r *http.Request) {
	var subject string
	idParam := chi.URLParam(r, "userID")
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

	if senderID < receiverID {
		subject = fmt.Sprintf("chat.%v.%v", senderID, receiverID)
	} else {
		subject = fmt.Sprintf("chat.%v.%v", receiverID, senderID)
	}
	err = app.nats.NatsConn.SendToChat(subject, bytePayload)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// since a text message is being sent, all the previous chats/posts are supposed to be displayed unlike in
	// case of sending a post where no previous chats/posts are supposed to be displayed
	if payload.Text != "" {
		fmt.Printf("Here\n")
		allChats, err := app.nats.NatsConn.GetallChats(subject)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		if err := app.jsonResponse(w, http.StatusOK, allChats); err != nil {
			app.internalServerError(w, r, err)
			return
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, ""); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getUserChat(w http.ResponseWriter, r *http.Request) {
	var subject string
	idParam := chi.URLParam(r, "userID")
	otherID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	loggedInUser := getUserFromContext(r)
	loggedInUserID := loggedInUser.ID

	if loggedInUserID < otherID {
		subject = fmt.Sprintf("chat.%v.%v", loggedInUserID, otherID)
	} else {
		subject = fmt.Sprintf("chat.%v.%v", otherID, loggedInUserID)
	}

	allChats, err := app.nats.NatsConn.GetallChats(subject)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, allChats); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
