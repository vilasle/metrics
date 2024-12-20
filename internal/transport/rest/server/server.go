package rest

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HTTPServer struct {
	srv     *http.Server
	stateMx *sync.RWMutex
	mux     *chi.Mux
	running bool
}

func NewHTTPServer(addr string) HTTPServer {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)

	return HTTPServer{
		srv: &http.Server{
			Addr:         addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		mux:     mux,
		stateMx: &sync.RWMutex{},
	}
}

func (s *HTTPServer) Register(path string, methods []string, contentTypes []string, handler http.HandlerFunc) {
	s.mux.Route(path, func(r chi.Router) {
		if len(methods) > 0 {
			r.Use(allowedMethods(methods...))
		}

		if len(contentTypes) > 0 {
			r.Use(allowedContentType(contentTypes...))
		}

		r.HandleFunc("/*", handler)
	})
}

func (s *HTTPServer) Start() error {
	s.srv.Handler = s.mux

	s.stateMx.Lock()
	s.running = true
	s.stateMx.Unlock()

	defer func() {
		s.stateMx.Lock()
		s.running = false
		s.stateMx.Unlock()
	}()

	err := s.srv.ListenAndServe()

	if err != nil && err == http.ErrServerClosed {
		err = nil
	}

	return err
}

func (s *HTTPServer) IsRunning() bool {
	s.stateMx.Lock()
	defer s.stateMx.Unlock()
	return s.running
}

func (s *HTTPServer) Stop() error {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		s.stateMx.Lock()
		s.running = false
		s.stateMx.Unlock()

		return err
	}
	return nil
}

func (s *HTTPServer) ForceStop() error {
	return s.srv.Close()
}
