package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swaggerui"
)

//go:embed index.html
var indexHTML []byte

// StockUpdate represents a stock price update event
type StockUpdate struct {
	Symbol    string    `json:"symbol" desc:"Stock ticker symbol"`
	Price     float64   `json:"price" desc:"Current stock price"`
	Change    float64   `json:"change" desc:"Price change since last update"`
	Timestamp time.Time `json:"timestamp" desc:"Time of the update"`
}

// ErrorEvent represents an error event
type ErrorEvent struct {
	Code    string `json:"code" desc:"Error code"`
	Message string `json:"message" desc:"Error message"`
}

// HeartbeatEvent represents a keep-alive event
type HeartbeatEvent struct {
	Timestamp time.Time `json:"timestamp" desc:"Server time"`
}

var stocks = []string{"AAPL", "MSFT", "GOOG", "AMZN", "META"}

func main() {
	r := router.New()

	generator := openapi.NewGenerator(openapi.Info{
		Title:       "SSE Example API",
		Version:     "1.0.0",
		Description: "Example API demonstrating Server-Sent Events with go-router",
	})

	r.GET("/", func(ctx *router.Context) {
		ctx.Writer.Header().Set("Content-Type", "text/html")
		ctx.Writer.Write(indexHTML)
	},
		openapi.WithSummary("Home Page"),
		openapi.WithDescription("Serves the home page with stock ticker demo"),
		openapi.ExcludeFromDocs(),
	)

	r.GET("/events/stocks", stockEventsHandler,
		openapi.WithTags("SSE"),
		openapi.WithSummary("Stock price updates stream"),
		openapi.WithSSEResponse("Real-time stock price updates via Server-Sent Events"),
		openapi.WithSSEEvent[StockUpdate]("stock_update", "Emitted when a stock price changes"),
		openapi.WithSSEEvent[HeartbeatEvent]("heartbeat", "Emitted periodically to keep the connection alive"),
		openapi.WithSSEEvent[ErrorEvent]("error", "Emitted when an error occurs"),
	)

	r.GET("/events/stocks/manual", stockEventsManual,
		openapi.WithTags("SSE"),
		openapi.WithSummary("Stock prices (manual handling)"),
		openapi.WithDescription("Same as /events/stocks but with manual SSE handling for comparison"),
		openapi.WithSSEResponse("Real-time stock price updates"),
		openapi.WithSSEEvent[StockUpdate]("stock_update", "Stock price changed"),
	)

	uiConfig := swaggerui.DefaultUIConfig()
	uiConfig.Title = "SSE Examples"
	uiConfig.DocExpansion = "list"

	setup := swaggerui.NewSetup(r, generator)
	setup.WithUIConfig(uiConfig)
	setup.RegisterRoutes(r, "/openapi.json", "/docs")

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	fmt.Println("Stock ticker demo at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// stockEventsHandler demonstrates the new SSEHandler wrapper.
// This is the recommended way to implement SSE endpoints.
func stockEventsHandler(c *router.Context) {
	stockPrices := make(map[string]float64)
	for _, symbol := range stocks {
		stockPrices[symbol] = 100.0 + float64(time.Now().Nanosecond()%2000)/100.0
	}

	c.SSEHandler(router.SSEConfig{
		KeepAliveInterval: 30 * time.Second,
		KeepAliveEnabled:  true,
		OnClientDisconnect: func() {
			log.Println("Client disconnected from stock events")
		},
	}, func(ctx context.Context, send router.SSESendFunc) error {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		heartbeatTicker := time.NewTicker(10 * time.Second)
		defer heartbeatTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil

			case <-ticker.C:
				symbol := stocks[time.Now().Unix()%int64(len(stocks))]
				change := (float64(time.Now().Nanosecond()%400) - 200.0) / 100.0
				stockPrices[symbol] += change

				update := StockUpdate{
					Symbol:    symbol,
					Price:     stockPrices[symbol],
					Change:    change,
					Timestamp: time.Now(),
				}

				if err := send("stock_update", update); err != nil {
					return err
				}

			case <-heartbeatTicker.C:
				if err := send("heartbeat", HeartbeatEvent{Timestamp: time.Now()}); err != nil {
					return err
				}
			}
		}
	})
}

// stockEventsManual demonstrates manual SSE handling (the old way).
// This is still supported for cases where you need more control.
func stockEventsManual(c *router.Context) {
	c.InitSSE()

	clientGone := c.Request.Context().Done()

	stockTicker := time.NewTicker(2 * time.Second)
	defer stockTicker.Stop()

	keepAliveTicker := time.NewTicker(15 * time.Second)
	defer keepAliveTicker.Stop()

	stockPrices := make(map[string]float64)
	for _, symbol := range stocks {
		stockPrices[symbol] = 100.0 + float64(time.Now().Nanosecond()%2000)/100.0
	}

	for {
		select {
		case <-clientGone:
			log.Println("Client disconnected from manual stock events")
			return

		case <-stockTicker.C:
			symbol := stocks[time.Now().Unix()%int64(len(stocks))]
			change := (float64(time.Now().Nanosecond()%400) - 200.0) / 100.0
			stockPrices[symbol] += change

			update := StockUpdate{
				Symbol:    symbol,
				Price:     stockPrices[symbol],
				Change:    change,
				Timestamp: time.Now(),
			}

			eventID := symbol + "-" + strconv.FormatInt(time.Now().Unix(), 10)
			if err := c.SSEJson("stock_update", update, eventID); err != nil {
				log.Printf("Error sending stock update: %v", err)
				return
			}

		case <-keepAliveTicker.C:
			if err := c.SSEKeepAlive(); err != nil {
				log.Printf("Error sending keep-alive: %v", err)
				return
			}
		}
	}
}
