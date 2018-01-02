package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Settings global variables is actually BAD practice
var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
var upgrader = websocket.Upgrader{}

// Message describes common chat message structure
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message

		// Read in new messages JSON and map them to Message struct
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	// Serve static assets
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configure WS route
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
