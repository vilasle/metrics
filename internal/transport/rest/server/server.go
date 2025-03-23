package rest

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mdw "github.com/vilasle/metrics/internal/transport/rest/middleware"
)

type HTTPServer struct {
	srv     *http.Server
	mux     *chi.Mux
	running atomic.Bool
}

func NewHTTPServer(addr string, middlewareOptions ...func(http.Handler) http.Handler) *HTTPServer {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.RequestID)
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

func (s *HTTPServer) Register(path string, methods []string, contentTypes []string, handler http.Handler) {
	s.mux.Route(path, func(r chi.Router) {
		if len(methods) > 0 {
			r.Use(mdw.AllowedMethods(methods...))
		}

		if len(contentTypes) > 0 {
			r.Use(mdw.AllowedContentType(contentTypes...))
		}
		r.Handle("/", handler)
	})
}

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

func (s *HTTPServer) IsRunning() bool {
	v := s.running.Load()
	return v
}

func (s *HTTPServer) Stop() error {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		s.running.Swap(false)
		return err
	}
	return nil
}

func (s *HTTPServer) ForceStop() error {
	return s.srv.Close()
}
