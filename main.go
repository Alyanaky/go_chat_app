package main

import (
    "html/template"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // связанные клиенты
var broadcast = make(chan Message)           // канал для широковещательной рассылки сообщений

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

// Message структура для хранения сообщений
type Message struct {
    Username string
    Message  string
}

func main() {
    fs := http.FileServer(http.Dir("static/"))
    http.Handle("/", fs)
    http.HandleFunc("/ws", handleConnections)

    go handleMessages()

    log.Println("Listening on :8080...")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
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
        err := ws.ReadJSON(&msg)
        if err != nil {
            delete(clients, ws) // удаление клиента при разрыве соединения
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
