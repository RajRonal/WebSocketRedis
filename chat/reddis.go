package chat

import (
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"log"
	"strings"
)

var client *redis.Client
var redisHost string
var redisPassword string
var sub *redis.PubSub

func init() {

	redisHost = "localhost:6379"
	if redisHost == "" {
		log.Fatal("missing REDIS_HOST env var")
	}

	redisPassword = ""

	log.Println("connecting to Redis...")
	client = redis.NewClient(&redis.Options{Addr: redisHost, Password: redisPassword, DB: 0})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal("failed to connect to redis", err)
	}
	log.Println("connected to redis", redisHost)
}

func startSubscriber(channelID string) {

	go func() {
		log.Println("starting subscriber...")
		sub = client.Subscribe(channelID)
		//error
		messages := sub.Channel()
		for message := range messages {
			from := strings.Split(message.Payload, ":")[0]
			for user, peer := range Peers {
				if from != user {
					peer.WriteMessage(websocket.TextMessage, []byte(message.Payload))
				}
			}
		}
	}()
}

func SendToChannelV2(msg, channel string) {
	err := client.Publish(channel, msg)
	if err != nil {
		log.Println("could not publish to channel", err)
	}
}

const users = "chat-users"

func CheckUserExists(user string) (bool, error) {
	usernameTaken, err := client.SIsMember(users, user).Result()
	if err != nil {
		return false, err
	}
	return usernameTaken, nil
}

func CreateUser(user string) error {
	err := client.SAdd(users, user).Err()
	if err != nil {
		return err
	}
	return nil
}

func RemoveUser(user string) {
	err := client.SRem(users, user).Err()
	if err != nil {
		log.Println("failed to remove user:", user)
		return
	}
	log.Println("removed user from redis:", user)
}
