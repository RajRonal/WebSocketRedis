package main

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"strings"
	"web-socket-redis/chat"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var redisHost string
var redisPassword string

const port = "8081"

func init() {
	redisHost = "localhost:6379"
	if redisHost == "" {
		logrus.Fatal("missing REDIS_HOST env var")
	}

	redisPassword = ""
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("you are good to go!"))
	})
	http.Handle("/ws/", http.HandlerFunc(WebSocketHandler))
	server := http.Server{Addr: ":" + port, Handler: nil}
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("failed to start server", err)
	}

}

func WebSocketHandler(writer http.ResponseWriter, request *http.Request) {
	user := strings.TrimPrefix(request.URL.Path, "/ws/")
	receiverID := request.URL.Query().Get("receiverID")

	peer, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Fatal("websocket conn failed", err)
	}

	chatSession := chat.NewChatSession(user, peer, receiverID)
	chatSession.Start(receiverID)
}
