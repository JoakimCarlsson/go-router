package router

import (
	"context"
	"time"
)

// SSEConfig configures Server-Sent Events behavior for an endpoint.
type SSEConfig struct {
	// KeepAliveInterval is the interval between keep-alive comments.
	// Set to 0 to disable automatic keep-alive.
	// Default: 15 seconds if not specified and KeepAliveEnabled is true.
	KeepAliveInterval time.Duration

	// KeepAliveEnabled controls whether automatic keep-alive is enabled.
	// If true and KeepAliveInterval is 0, defaults to 15 seconds.
	KeepAliveEnabled bool

	// OnClientDisconnect is called when the client disconnects.
	// This is useful for cleanup operations like closing channels or
	// releasing resources associated with the connection.
	OnClientDisconnect func()
}

// DefaultSSEConfig returns a default SSE configuration with keep-alive enabled.
func DefaultSSEConfig() SSEConfig {
	return SSEConfig{
		KeepAliveInterval: 15 * time.Second,
		KeepAliveEnabled:  true,
	}
}

// SSESendFunc is the function signature for sending SSE events.
// It sends an event with the given name and data to the client.
// Returns an error if the send fails (e.g., client disconnected).
type SSESendFunc func(event string, data any) error

// SSEHandlerFunc is the function signature for SSE handler logic.
// The context is cancelled when the client disconnects.
// The send function should be used to emit events to the client.
type SSEHandlerFunc func(ctx context.Context, send SSESendFunc) error

// SSEHandler wraps SSE endpoint logic with automatic initialization,
// keep-alive management, and client disconnection handling.
//
// The handler function receives a context that is cancelled when the
// client disconnects, and a send function to emit events.
//
// Example:
//
//	func streamHandler(c *router.Context) {
//	    c.SSEHandler(router.DefaultSSEConfig(), func(ctx context.Context, send router.SSESendFunc) error {
//	        ticker := time.NewTicker(2 * time.Second)
//	        defer ticker.Stop()
//
//	        for {
//	            select {
//	            case <-ctx.Done():
//	                return nil
//	            case <-ticker.C:
//	                if err := send("update", myData); err != nil {
//	                    return err
//	                }
//	            }
//	        }
//	    })
//	}
func (c *Context) SSEHandler(config SSEConfig, handler SSEHandlerFunc) {
	c.InitSSE()

	ctx := c.Request.Context()

	send := func(event string, data any) error {
		return c.SSEJson(event, data, "")
	}

	if config.KeepAliveEnabled {
		interval := config.KeepAliveInterval
		if interval == 0 {
			interval = 15 * time.Second
		}

		keepAliveCtx, cancelKeepAlive := context.WithCancel(ctx)
		defer cancelKeepAlive()

		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-keepAliveCtx.Done():
					return
				case <-ticker.C:
					if err := c.SSEKeepAlive(); err != nil {
						return
					}
				}
			}
		}()

		err := handler(ctx, send)

		cancelKeepAlive()

		if err != nil || ctx.Err() != nil {
			if config.OnClientDisconnect != nil {
				config.OnClientDisconnect()
			}
		}
	} else {
		err := handler(ctx, send)
		if err != nil || ctx.Err() != nil {
			if config.OnClientDisconnect != nil {
				config.OnClientDisconnect()
			}
		}
	}
}

// SSEHandlerSimple is a simplified version of SSEHandler that uses default configuration.
// It provides automatic keep-alive every 15 seconds and handles client disconnection.
//
// Example:
//
//	func streamHandler(c *router.Context) {
//	    c.SSEHandlerSimple(func(ctx context.Context, send router.SSESendFunc) error {
//	        for {
//	            select {
//	            case <-ctx.Done():
//	                return nil
//	            case msg := <-messages:
//	                if err := send("message", msg); err != nil {
//	                    return err
//	                }
//	            }
//	        }
//	    })
//	}
func (c *Context) SSEHandlerSimple(handler SSEHandlerFunc) {
	c.SSEHandler(DefaultSSEConfig(), handler)
}
