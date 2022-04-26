package hub

import (
	"log"
	"net/http"
)

type Server struct {
	cfg               *Config
	logger            *log.Logger
	webHookDispatcher *Dispatcher
}

func NewServer(cfg *Config, logger *log.Logger, dispatcher *Dispatcher) *Server {
	return &Server{
		cfg:               cfg,
		logger:            logger,
		webHookDispatcher: dispatcher,
	}
}

func (s *Server) Start() error {
	addr := s.cfg.Addr()
	mux := http.NewServeMux()
	mux.Handle("/", s.webHookDispatcher)
	srv := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	s.logger.Printf("start server on %s", addr)
	return srv.ListenAndServe()
}
