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

func CarShowRouteOptions() []router.RouteOption {
	return []router.RouteOption{
		docs.WithTags("Car Show"),
		docs.WithSummary("Get car show information"),
		docs.WithDescription("Retrieves detailed car show information with featured cars, manufacturers, and complex relationships"),
		docs.WithQueryParam("showId", "string", true, "Car Show ID to retrieve", "classic-cars-2024"),
		docs.WithQueryParam("offset", "integer", false, "Offset for pagination", 0),
		docs.WithQueryParam("limit", "integer", false, "Limit for pagination", 10),
		docs.WithQueryParam("filterType", "string", false, "Filter type (vintage, modern, electric)", "vintage"),
		docs.WithResponse(200, "Car show retrieved successfully"),
		docs.WithJSONResponse[CarShowResponse](200, "Complete car show information with complex circular references"),
		docs.WithResponse(400, "Bad request - invalid parameters"),
		docs.WithResponse(404, "Car show not found"),
		docs.WithResponse(500, "Internal server error"),
	}
}

func GetCarShow(ctx *router.Context) {
	showID := ctx.Query().Get("showId")
	if showID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "showId is required",
		})
		return
	}

	offset := ctx.QueryIntDefault("offset", 0)
	limit := ctx.QueryIntDefault("limit", 10)
	filterType := ctx.Query().Get("filterType")

	response := createMockCarShowResponse(showID, offset, limit, filterType)

	ctx.JSON(http.StatusOK, response)
}

func createMockCarShowResponse(showID string, offset, limit int, filterType string) CarShowResponse {
	// Create some cars first
	ferrari := Car{
		ID:        "ferrari-488",
		Make:      "Ferrari",
		Model:     "488 GTB",
		Year:      2020,
		Color:     "Rosso Corsa Red",
		IsVintage: false,
		Price:     280000,
		Engine: &Engine{
			Type:         "V8 Twin-Turbo",
			Displacement: 3.9,
			Horsepower:   661,
			Torque:       561,
			FuelType:     "Premium Gasoline",
			Cylinders:    8,
		},
	}

	porsche := Car{
		ID:        "porsche-911",
		Make:      "Porsche",
		Model:     "911 Carrera",
		Year:      2023,
		Color:     "Guards Red",
		IsVintage: false,
		Price:     115000,
		Engine: &Engine{
			Type:         "Flat-6 Twin-Turbo",
			Displacement: 3.0,
			Horsepower:   379,
			Torque:       331,
			FuelType:     "Premium Gasoline",
			Cylinders:    6,
		},
	}

	// Create manufacturers with circular refs
	ferrariMfg := Manufacturer{
		ID:      "ferrari-spa",
		Name:    "Ferrari S.p.A.",
		Country: "Italy",
		Founded: 1947,
		Models:  []Car{ferrari},
	}

	porscheMfg := Manufacturer{
		ID:      "porsche-ag",
		Name:    "Dr. Ing. h.c. F. Porsche AG",
		Country: "Germany",
		Founded: 1931,
		Models:  []Car{porsche},
	}

	// Add cross-references
	ferrari.Manufacturer = &ferrariMfg
	porsche.Manufacturer = &porscheMfg
	ferrari.RelatedCars = []Car{porsche}
	porsche.RelatedCars = []Car{ferrari}

	// Create the main car show
	mainShow := CarShow{
		ID:           showID,
		Name:         "Classic & Exotic Car Showcase 2024",
		Location:     "Monaco Convention Center",
		Date:         "2024-09-15",
		Description:  "The ultimate gathering of automotive excellence featuring rare classics and modern supercars",
		FeaturedCars: []Car{ferrari, porsche},
		IsActive:     true,
		MaxCapacity:  5000,
		TicketPrice:  150.0,
		Theme:        "Speed & Elegance",
	}

	// Create related shows with circular references
	relatedShow := CarShow{
		ID:           "pebble-beach-2024",
		Name:         "Pebble Beach Concours d'Elegance",
		Location:     "Pebble Beach, California",
		Date:         "2024-08-20",
		Description:  "America's premier automotive showcase",
		FeaturedCars: []Car{ferrari},
		RelatedShows: []CarShow{mainShow},
		IsActive:     true,
		MaxCapacity:  3000,
		TicketPrice:  200.0,
		Theme:        "Automotive Art",
	}

	mainShow.RelatedShows = []CarShow{relatedShow}
	mainShow.ChildShows = []CarShow{relatedShow}

	// Add shows to cars
	ferrari.Shows = []CarShow{mainShow, relatedShow}
	porsche.Shows = []CarShow{mainShow}

	// Create user profile with circular references
	userProfile := UserProfile{
		ID:            "user-123",
		Username:      "car_enthusiast_42",
		Email:         "carfan@example.com",
		FavoriteCars:  []Car{ferrari, porsche},
		FavoriteShows: []CarShow{mainShow, relatedShow},
		CreatedAt:     "2023-01-15T10:30:00Z",
		LastActive:    "2024-08-25T14:22:00Z",
	}

	// Add user preferences with circular ref back to user
	userProfile.Preferences = &UserPreferences{
		FavoriteMakes: []string{"Ferrari", "Porsche", "Lamborghini"},
		MaxPrice:      500000,
		User:          &userProfile,
	}

	// Add friends (circular user references)
	friend := UserProfile{
		ID:            "user-456",
		Username:      "speed_demon",
		Email:         "speedy@example.com",
		FavoriteCars:  []Car{ferrari},
		FavoriteShows: []CarShow{mainShow},
		Friends:       []UserProfile{userProfile},
		CreatedAt:     "2022-11-20T09:15:00Z",
		LastActive:    "2024-08-24T16:45:00Z",
	}
	userProfile.Friends = []UserProfile{friend}

	return CarShowResponse{
		Data: CarShowData{
			CarShow:       mainShow,
			FeaturedCars:  []Car{ferrari, porsche},
			Manufacturers: []Manufacturer{ferrariMfg, porscheMfg},
		},
		Metadata: ResponseMetadata{
			RequestID:    "req-" + showID + "-" + fmt.Sprintf("%d", offset),
			Timestamp:    "2024-08-25T17:30:00Z",
			RelatedShows: []CarShow{relatedShow},
			UserProfile:  &userProfile,
		},
		Relations: &CarRelations{
			CompetingCars: []Car{ferrari, porsche},
			SimilarShows:  []CarShow{relatedShow},
			UserFavorites: &userProfile,
		},
	}
}

func main() {
	r := router.New()

	r.GET("/api/v1/carshow", GetCarShow, CarShowRouteOptions()...)

	// Create OpenAPI generator
	generator := openapi.NewGenerator(metadata.Info{
		Title:       "Car Show API - Complex Circular References Demo",
		Version:     "1.0.0",
		Description: "A fun car show API demonstrating complex JSON structures with circular references handled by go-router",
	})

	// Configure Swagger UI with specific settings for complex nested types
	uiConfig := swagger.DefaultUIConfig()
	uiConfig.DefaultModelRendering = "example"
	uiConfig.Title = "🏎️ Car Show API - Circular References Demo"
	uiConfig.DefaultModelsExpandDepth = 3

	// Set up the integration
	swaggerUI := integration.NewSwaggerUIIntegration(r, generator)
	swaggerUI.WithUIConfig(uiConfig)
	swaggerUI.SetupRoutes(r, "/openapi.json", "/docs")

	fmt.Println("API documentation available at: http://localhost:8080/docs")
	fmt.Println("OpenAPI spec available at: http://localhost:8080/openapi.json")

	log.Fatal(http.ListenAndServe(":8088", r))
}
