package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// StockUpdate represents a stock price update
type StockUpdate struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Change    float64   `json:"change"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	stocks = []string{"AAPL", "MSFT", "GOOG", "AMZN", "META"}
)

func main() {
	r := router.New()

	// Create OpenAPI generator
	generator := openapi.NewGenerator(metadata.Info{
		Title:       "SSE Example API",
		Version:     "1.0.0",
		Description: "Example API demonstrating Server-Sent Events with go-router",
	})

	r.GET("/", func(ctx *router.Context) {
		content, err := os.ReadFile("index.html")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		}
		ctx.Writer.Header().Set("Content-Type", "text/html")
		ctx.Writer.Write(content)
	},
		docs.WithSummary("Home Page"),
		docs.WithDescription("Serves the home page with chat UI and stock ticker demo"),
	)

	r.GET("/events/stocks", stockEvents,
		docs.WithTags("SSE"),
		docs.WithSummary("Stock price updates"),
		docs.WithDescription("Server-sent events stream for stock price updates"),
	)
	// Configure Swagger UI
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.Title = "SSE Examples"
	uiConfig.DocExpansion = "list"

	// Set up Swagger integration
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	// Start the server
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API documentation available at http://localhost:8080/docs")
	fmt.Println("Chat demo available at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// stockEvents streams simulated stock price updates as SSE
func stockEvents(c *router.Context) {
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

	stockChanges := make(map[string]float64)

	for {
		select {
		case <-clientGone:
			log.Println("Client disconnected from stock events")
			return

		case <-stockTicker.C:
			symbol := stocks[time.Now().Unix()%int64(len(stocks))]

			change := (float64(time.Now().Nanosecond()%400) - 200.0) / 100.0
			stockPrices[symbol] += change
			stockChanges[symbol] = change

			update := StockUpdate{
				Symbol:    symbol,
				Price:     stockPrices[symbol],
				Change:    change,
				Timestamp: time.Now(),
			}

			err := c.SSEJson("stock_update", update, symbol+"-"+strconv.FormatInt(time.Now().Unix(), 10))
			if err != nil {
				log.Printf("Error sending stock update: %v", err)
				return
			}

		case <-keepAliveTicker.C:
			err := c.SSEKeepAlive()
			if err != nil {
				log.Printf("Error sending keep-alive: %v", err)
				return
			}
		}
	}
}
