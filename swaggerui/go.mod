module github.com/joakimcarlsson/go-router/swaggerui

go 1.22

require (
	github.com/joakimcarlsson/go-router/openapi v1.0.1
	github.com/joakimcarlsson/go-router/router/v2 v2.0.1
)

replace (
	github.com/joakimcarlsson/go-router/openapi => ../openapi
	github.com/joakimcarlsson/go-router/router/v2 => ../router
)
