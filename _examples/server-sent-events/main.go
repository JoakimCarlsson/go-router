package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/joakimcarlsson/go-router/router"
)

//go:embed index.html
var indexHTML []byte

type Message struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type ServerTime struct {
	Time   string `json:"time"`
	Unix   int64  `json:"unix"`
	Uptime string `json:"uptime"`
}

var serverStart = time.Now()

func main() {
	r := router.New()

	r.GET("/", func(c *router.Context) {
		c.Writer.Header().Set("Content-Type", "text/html")
		c.Writer.Write(indexHTML)
	})

	r.GET("/events/time", timeStreamHandler)
	r.GET("/events/messages", messageStreamHandler)
	r.GET("/events/manual", manualStreamHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func timeStreamHandler(c *router.Context) {
	c.SSEHandler(router.SSEConfig{
		KeepAliveInterval: 30 * time.Second,
		KeepAliveEnabled:  true,
	}, func(ctx context.Context, send router.SSESendFunc) error {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		if err := send("time", ServerTime{
			Time:   time.Now().Format(time.RFC3339),
			Unix:   time.Now().Unix(),
			Uptime: time.Since(serverStart).Round(time.Second).String(),
		}); err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := send("time", ServerTime{
					Time:   time.Now().Format(time.RFC3339),
					Unix:   time.Now().Unix(),
					Uptime: time.Since(serverStart).Round(time.Second).String(),
				}); err != nil {
					return err
				}
			}
		}
	})
}

func messageStreamHandler(c *router.Context) {
	messages := []string{
		"Welcome to the SSE demo!",
		"Server-Sent Events are great for real-time updates",
		"Perfect for notifications and live feeds",
	}

	c.SSEHandlerSimple(func(ctx context.Context, send router.SSESendFunc) error {
		messageID := 0
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				if err := send("message", Message{
					ID:        messageID,
					Text:      messages[messageID%len(messages)],
					Timestamp: time.Now(),
				}); err != nil {
					return err
				}
				messageID++
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(3 * time.Second):
				}
			}
		}
	})
}

func manualStreamHandler(c *router.Context) {
	c.InitSSE()

	clientGone := c.Request.Context().Done()
	eventTicker := time.NewTicker(2 * time.Second)
	defer eventTicker.Stop()

	keepAliveTicker := time.NewTicker(15 * time.Second)
	defer keepAliveTicker.Stop()

	counter := 0
	for {
		select {
		case <-clientGone:
			return
		case <-eventTicker.C:
			counter++
			if err := c.SSEJson("counter", map[string]interface{}{
				"count": counter,
				"time":  time.Now().Format(time.RFC3339),
			}, strconv.Itoa(counter)); err != nil {
				return
			}
		case <-keepAliveTicker.C:
			if err := c.SSEKeepAlive(); err != nil {
				return
			}
		}
	}
}
