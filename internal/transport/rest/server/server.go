package rest

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// HTTPServer is the structure that holds and wraps the http server
type HTTPServer struct {
	srv     *http.Server
	mux     *chi.Mux
	running atomic.Bool
}

// NewHTTPServer create new instance of HTTPServer
// addr is the address to listen on
// middlewareOptions are the middleware to use
func NewHTTPServer(addr string, middlewareOptions ...func(http.Handler) http.Handler) *HTTPServer {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	for _, m := range middlewareOptions {
		mux.Use(m)
	}

	mux.Mount("/debug", middleware.Profiler())

	srv := &HTTPServer{
		srv: &http.Server{
			Addr:         addr,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		mux:     mux,
		running: atomic.Bool{},
	}
	srv.running.Store(false)

	return srv
}

// Register - register new handlers
func (s *HTTPServer) Register(path string, handler http.Handler, methods ...string) {
	if len(methods) == 0 {
		s.mux.Handle(path, handler)
		return
	}

	for _, method := range methods {
		s.mux.Method(method, path, handler)
	}
}

// Start - start the server
func (s *HTTPServer) Start() error {
	s.srv.Handler = s.mux
	s.running.Swap(true)

	defer s.running.Swap(false)

	err := s.srv.ListenAndServe()

	if err != nil && err == http.ErrServerClosed {
		err = nil
	}

	return err
}

// IsRunning - check if the server is running
func (s *HTTPServer) IsRunning() bool {
	v := s.running.Load()
	return v
}

// Stop - tries to stop the server gracefully
func (s *HTTPServer) Stop(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		s.running.Swap(false)
		return err
	}
	return nil
}

// ForceStop - tries to stop the server forcefully
func (s *HTTPServer) ForceStop() error {
	return s.srv.Close()
}
