package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[string]*websocket.Conn)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	var (
		target  string
		members []string
		// targetID int64
	)
	client, _ := strconv.Atoi(r.URL.Query().Get("clientid"))
	clientID := int64(client)
	target = r.URL.Query().Get("targetid")
	// if target != "" {
	// 	targetID, _ = strconv.ParseInt(target, 10, 64)
	// }

	members = strings.Split(r.URL.Query().Get("members"), ",")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrade error: %v\n", err)
		return
	}

	if target != "" {
		clients[r.URL.Query().Get("clientid")] = conn
	} else if len(members) != 0 {
		unique_ws := fmt.Sprintf("%v-%v", members[0], r.URL.Query().Get("clientid")) // members[0] is groupID
		clients[unique_ws] = conn
	}
	fmt.Printf("%v connected\n", clientID)

	go func() {
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("Read error: %v\n", err)
				delete(clients, r.URL.Query().Get("clientid"))
				break
			}

			if target != "" {
				targetConn, ok := clients[target]
				if ok {
					err = targetConn.WriteMessage(websocket.TextMessage, msg)
					if err != nil {
						fmt.Printf("Write error: %v\n", err)
					}
				}
			} else if len(members) != 0 {
				for _, member := range members {
					// targetID, _ = strconv.ParseInt(member, 10, 64)
					unique_ws := fmt.Sprintf("%v-%v", members[0], member)
					targetConn, ok := clients[unique_ws]
					if ok {
						err = targetConn.WriteMessage(websocket.TextMessage, msg)
						if err != nil {
							fmt.Printf("Write error: %v\n", err)
						}
					}
				}
			}
		}
	}()
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Server started on :5500")
	http.ListenAndServe(":5500", nil)
}
