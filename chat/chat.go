package chat

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
)

const usernameHasBeenTaken = "username %s is already taken. please retry with a different name"
const retryMessage = "failed to connect. please try again"
const welcome = "Welcome %s!"
const joined = "%s: has joined the chat!"
const chat = "%s: %s"
const left = "%s: has left the chat!"

var Peers map[string]*websocket.Conn

func init() {
	Peers = map[string]*websocket.Conn{}
}

type ChatSession struct {
	user       string
	peer       *websocket.Conn
	receiverID string
	channelID  string
}

func NewChatSession(user string, peer *websocket.Conn, receiverID string) *ChatSession {
	userid, _ := strconv.Atoi(user)
	receiverIDInt, _ := strconv.Atoi(receiverID)
	var channelID string
	if userid > receiverIDInt {
		channelID = fmt.Sprintf("%v##%v", receiverIDInt, userid)
		fmt.Println(channelID)
	} else {
		channelID = fmt.Sprintf("%v##%v", userid, receiverIDInt)
		fmt.Println(channelID)

	}
	startSubscriber(channelID)
	return &ChatSession{user: user, peer: peer, receiverID: receiverID, channelID: channelID}
}

func (s *ChatSession) Start(receiverId string) {
	usernameTaken, err := CheckUserExists(s.user)

	if err != nil {
		log.Println("unable to determine whether user exists -", s.user)
		s.notifyPeer(retryMessage)
		s.peer.Close()
		return
	}

	if usernameTaken {
		msg := fmt.Sprintf(usernameHasBeenTaken, s.user)
		s.peer.WriteMessage(websocket.TextMessage, []byte(msg))
		s.peer.Close()
		return
	}

	err = CreateUser(s.user)
	if err != nil {
		log.Println("failed to add user to list of active chat users", s.user)
		s.notifyPeer(retryMessage)
		s.peer.Close()
		return
	}
	Peers[s.user] = s.peer

	s.notifyPeer(fmt.Sprintf(welcome, s.user))
	SendToChannelV2(fmt.Sprintf(joined, s.user), s.channelID)

	go func() {
		log.Println("user joined", s.user)
		for {
			//_, ok := Peers[receiverId]
			//if ok {
			_, msg, err := s.peer.ReadMessage()
			if err != nil {
				_, ok := err.(*websocket.CloseError)
				if ok {
					log.Println("connection closed by user")
					s.disconnect()
				}
				return
			}
			//_, ok := Peers[receiverId]
			//if ok {
			SendToChannelV2(fmt.Sprintf(chat, s.user, string(msg)), s.channelID)
			//}

		}
	}()
}

func (s *ChatSession) notifyPeer(msg string) {
	err := s.peer.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("failed to write message", err)
	}
}

func (s *ChatSession) disconnect() {
	//remove user from SET
	RemoveUser(s.user)
	SendToChannelV2(fmt.Sprintf(left, s.user), s.channelID)
	s.peer.Close()
	delete(Peers, s.user)
}
