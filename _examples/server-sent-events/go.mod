module example/sse

go 1.22

require (
	github.com/joakimcarlsson/go-router/integration v0.0.0
	github.com/joakimcarlsson/go-router/openapi v0.0.0
	github.com/joakimcarlsson/go-router/router v0.0.0
	github.com/joakimcarlsson/go-router/swaggerui v0.0.0
)

replace (
	github.com/joakimcarlsson/go-router/integration => ../../integration
	github.com/joakimcarlsson/go-router/openapi => ../../openapi
	github.com/joakimcarlsson/go-router/router => ../../router
	github.com/joakimcarlsson/go-router/swaggerui => ../../swaggerui
)
