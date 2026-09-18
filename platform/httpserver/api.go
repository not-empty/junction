package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type Middleware func(next http.Handler) http.Handler

type APIServer struct {
	addr        string
	router      http.Handler
	middlewares []Middleware
}

func NewAPIServer(addr string, customRouter http.Handler) *APIServer {
	return &APIServer{
		addr:   addr,
		router: customRouter,
	}
}

func (s *APIServer) UseMiddleware(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

func (s *APIServer) Run(ctx context.Context) error {
	middlewareChain := applyMiddlewareChain(s.middlewares...)

	server := http.Server{
		Addr:              s.addr,
		Handler:           middlewareChain(s.router),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server started", "addr", s.addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

func applyMiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next
	}
}
