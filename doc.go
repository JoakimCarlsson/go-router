// Package router provides a lightweight HTTP router with OpenAPI support.
//
// Example usage:
//
//	r := router.New()
//	r.GET("/users", handleUsers)
//	r.POST("/users", createUser)
//	http.ListenAndServe(":8080", r)
package router