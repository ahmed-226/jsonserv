package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"myserv/internal/config"
	"myserv/internal/handlers"
	"myserv/internal/store"
)

type Server struct {
	config  *config.Config
	store   *store.Store
	handler http.Handler
}

func New(cfg *config.Config, s *store.Store) *Server {
	mux := http.NewServeMux()
	h := handlers.New(s, cfg)

	for _, entity := range s.Entities() {
		ecfg := cfg.EntityConfig(entity)
		path := entity
		if ecfg.Alias != "" {
			path = ecfg.Alias
		}

		mux.HandleFunc("GET /"+path, h.List(entity))
		mux.HandleFunc("GET /"+path+"/{id}", h.Get(entity))
		mux.HandleFunc("POST /"+path, h.Create(entity))
		mux.HandleFunc("PUT /"+path+"/{id}", h.Update(entity))
		mux.HandleFunc("PATCH /"+path+"/{id}", h.Patch(entity))
		mux.HandleFunc("DELETE /"+path+"/{id}", h.Delete(entity))

		log.Printf("  route: %s -> /%s", entity, path)
	}

	var handler http.Handler = mux
	handler = recoveryMiddleware(handler)
	if cfg.Server.Cors.Enabled {
		handler = corsMiddleware(handler, cfg.Server.Cors.Origins)
	}
	if cfg.Server.Logging {
		handler = loggingMiddleware(handler)
	}

	return &Server{config: cfg, store: s, handler: handler}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("jsonserv running on http://%s", addr)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      s.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return httpServer.ListenAndServe()
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler, origins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, o := range origins {
			if o == "*" || o == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
