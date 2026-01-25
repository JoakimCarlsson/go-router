module example/output-cache

go 1.22

require (
	github.com/joakimcarlsson/go-router/outputcache v0.0.0
	github.com/joakimcarlsson/go-router/router v0.0.0
)

replace (
	github.com/joakimcarlsson/go-router/outputcache => ../../outputcache
	github.com/joakimcarlsson/go-router/router => ../../router
)
