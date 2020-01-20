package http

import (
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// DefaultAddr is the default bind address.
const DefaultAddr = ":8000"

// Server represents an HTTP server.
type Server struct {
	ln   net.Listener
	Addr string
	log  *logrus.Logger
}

// NewServer returns a new instance of Server.
func NewServer(log *logrus.Logger) *Server {
	if log == nil {
		log = logrus.New()
	}

	return &Server{
		Addr: DefaultAddr,
		log:  log,
	}
}

// Open opens a socket and serves.
func (s *Server) Open() error {
	s.log.WithField("Bind address", s.Addr).Info("Binding address")
	ln, err := net.Listen("tcp", s.Addr)
	s.log.WithField("Bind address", s.Addr).Info("Listening at")
	if err != nil {
		return err
	}
	s.ln = ln
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(WithKubeClient())
	r.Mount("/api/v1", s.getAPIRouter())
	err = http.Serve(s.ln, r)
	if err != nil {
		return err
	}
	return nil
}

// Close closes all services and database.
func (s *Server) Close() error {
	// Close listener.
	if s.ln != nil {
		s.ln.Close()
	}
	s.log.Info("Server closed")
	return nil
}
