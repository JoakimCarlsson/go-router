package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joakimcarlsson/go-router/docs"
	"github.com/joakimcarlsson/go-router/integration"
	"github.com/joakimcarlsson/go-router/metadata"
	"github.com/joakimcarlsson/go-router/openapi"
	"github.com/joakimcarlsson/go-router/router"
	"github.com/joakimcarlsson/go-router/swagger"
)

func PanelRouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("Panel"),
		docs.WithSummary("Get panel information"),
		docs.WithDescription("Retrieves panel information and content by panel ID with pagination support"),
		docs.WithQueryParam("panelId", "string", true, "Panel ID to retrieve", "3DkAXJqTwajM7RytxM0ju7"),
		docs.WithQueryParam("offset", "integer", false, "Offset for pagination", 0),
		docs.WithQueryParam("limit", "integer", false, "Limit for pagination", 16),
		docs.WithQueryParam("skipProgress", "boolean", false, "Whether to skip progress information", true),
		docs.WithResponse(200, "Panel retrieved successfully"),
		docs.WithJSONResponse[PanelResponse](200, "Panel information with content and pagination"),
		docs.WithResponse(400, "Bad request - invalid parameters"),
		docs.WithResponse(500, "Internal server error"),
	}
}

func GetPanel(ctx *router.Context) {
	panelID := ctx.Query().Get("panelId")
	if panelID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "panelId is required",
		})
		return
	}

	offset := ctx.QueryIntDefault("offset", 0)
	limit := ctx.QueryIntDefault("limit", 16)
	skipProgress := ctx.QueryBoolDefault("skipProgress", false)

	response := createMockPanelResponse(panelID, offset, limit, skipProgress)

	ctx.JSON(http.StatusOK, response)
}

func createMockPanelResponse(panelID string, offset, limit int, skipProgress bool) PanelResponse {
	return PanelResponse{
		Data: PanelData{
			Panel: Panel{
				Typename:    "Panel",
				ID:          panelID,
				Title:       "Featured Movies",
				DisplayHint: &DisplayHint{MediaPanelImageRatio: "16:9"},
				Content: &PanelContent{
					PageInfo: PageInfo{
						HasNextPage:    true,
						NextPageOffset: offset + limit,
						TotalCount:     100,
					},
					Cards: []PanelCard{
						{
							Typename: "MovieCard",
							Movie: &Movie{
								Typename:            "Movie",
								ID:                  "movie-123",
								Slug:                "action-hero-2024",
								Title:               "Action Hero",
								HumanCallToAction:   "Watch Now",
								Genres:              []string{"Action", "Adventure"},
								ProductionYear:      "2024",
								MediaClassification: "Movie",
								Synopsis: &Synopsis{
									Brief:  "An action-packed adventure",
									Medium: "A thrilling story of heroism and adventure in modern times",
									Long:   "Follow our protagonist as they navigate through dangerous situations to save the world from an imminent threat.",
								},
								Images: &MovieImages{
									Main16x9: &Image{
										ID:     "img-123",
										Source: "https://example.com/movie-poster.jpg",
									},
									Cover2x3: &Image{
										ID:     "img-124",
										Source: "https://example.com/movie-cover.jpg",
									},
								},
								Duration: &Duration{
									ReadableShort: "2h 15m",
									Seconds:       8100,
								},
								Access: &Access{HasAccess: true},
							},
						},
						{
							Typename: "SeriesCard",
							Series: &Series{
								Typename:                 "Series",
								ID:                       "series-456",
								Slug:                     "mystery-drama-2024",
								Title:                    "Mystery Drama",
								Genres:                   []string{"Drama", "Mystery"},
								MediaClassification:      "Series",
								NumberOfAvailableSeasons: 3,
								Synopsis: &Synopsis{
									Brief:  "A compelling mystery series",
									Medium: "Uncover secrets in this gripping drama series",
									Long:   "A detective investigates a series of mysterious events that shake a small town to its core.",
								},
								Images: &SeriesImages{
									Main16x9: &Image{
										ID:     "img-789",
										Source: "https://example.com/series-poster.jpg",
									},
									Cover2x3: &Image{
										ID:     "img-790",
										Source: "https://example.com/series-cover.jpg",
									},
								},
							},
						},
					},
				},
				Pitch:      "Discover amazing content",
				ShortPitch: "Great entertainment awaits",
				LinkText:   "View All",
				Images: &PanelImages{
					Image16x9: &Image{
						ID:     "panel-img-1",
						Source: "https://example.com/panel-bg.jpg",
					},
				},
			},
		},
	}
}

func main() {
	r := router.New()

	r.GET("/api/v1/panel", GetPanel, PanelRouteOptions()...)

	// Create OpenAPI generator
	generator := openapi.NewGenerator(metadata.Info{
		Title:       "Complex JSON Panel API",
		Version:     "1.0.0",
		Description: "A sample panel API demonstrating complex JSON structures with go-router",
	})

	// Configure Swagger UI with specific settings for complex nested types
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.DefaultModelRendering = "example"
	uiConfig.Title = "Complex JSON Panel API"
	uiConfig.DefaultModelsExpandDepth = 2

	// Set up the integration
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Try: http://localhost:8080/api/v1/panel?panelId=3DkAXJqTwajM7RytxM0ju7&offset=0&limit=5")
	fmt.Println("API documentation available at: http://localhost:8080/docs")
	fmt.Println("OpenAPI spec available at: http://localhost:8080/openapi.json")

	log.Fatal(http.ListenAndServe(":8080", r))
}
