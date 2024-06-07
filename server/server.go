package server

import (
	"log/slog"
	"net/http"

	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/radarr"
	"github.com/flxp49/notion-watchlistarr/internal/sonarr"
)

type Server struct {
	listenAddr string
	N          *notion.NotionClient
	R          *radarr.RadarrClient
	S          *sonarr.SonarrClient
	Logger     *slog.Logger
	RadarrInit bool
	SonarrInit bool
}

func NewServer(listenAddr string, N *notion.NotionClient, R *radarr.RadarrClient, S *sonarr.SonarrClient, Logger *slog.Logger, RadarrInit bool, SonarrInit bool) *Server {
	return &Server{
		listenAddr: listenAddr,
		N:          N,
		R:          R,
		S:          S,
		Logger:     Logger,
		RadarrInit: RadarrInit,
		SonarrInit: SonarrInit,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.incorrectReqHandler)
	if s.RadarrInit {
		http.HandleFunc("POST /radarr", s.radarrHandler)
	}
	if s.SonarrInit {
		http.HandleFunc("POST /sonarr", s.sonarrHandler)
	}
	return http.ListenAndServe(":"+s.listenAddr, nil)
}
