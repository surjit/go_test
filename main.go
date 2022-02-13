package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"test4/controllers"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var ctx = context.Background()

var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func TestSocket() fiber.Handler {
	socket := websocket.New(func(c *websocket.Conn) {

		go deliverMessages(c)

		var (
			msg []byte
			err error
		)
		for {
			if _, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}

			if err := redisClient.Publish(ctx, "chat", msg).Err(); err != nil {
				log.Println("publish:", err)
				break
			}
		}
	})

	return socket
}

func deliverMessages(c *websocket.Conn) {
	subscriber := redisClient.Subscribe(ctx, "chat")
	user := User{}

	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			log.Println("subscriber:", err)
			panic(err)
		}

		if err := json.Unmarshal([]byte(msg.Payload), &user); err != nil {
			log.Println("Unmarshal:", err)
			panic(err)
		}

		text := []byte(fmt.Sprintf("{\"name\":\"%s\", \"email\":\"%s\"}", user.Name, user.Email))
		if err = c.WriteMessage(websocket.TextMessage, text); err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	app := fiber.New(fiber.Config{
		Prefork:               true,
		CaseSensitive:         true,
		StrictRouting:         true,
		DisableStartupMessage: true,
		ServerHeader:          "Test v3",
	})

	// Get all records from MySQL
	app.Get("/", controllers.Home)

	app.Get("/ws", TestSocket())

	log.Fatal(app.Listen("0.0.0.0:3000"))
}
