package api

import (
	"log/slog"
	"net/http"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/sonarr"
)

type Server struct {
	listenAddr string
	N          *notion.NotionClient
	R          *radarr.RadarrClient
	S          *sonarr.SonarrClient
	Logger     *slog.Logger
}

func NewServer(listenAddr string, N *notion.NotionClient, R *radarr.RadarrClient, S *sonarr.SonarrClient, Logger *slog.Logger) *Server {
	return &Server{
		listenAddr: listenAddr,
		N:          N,
		R:          R,
		S:          S,
		Logger:     Logger,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.incorrectReqHandler)
	http.HandleFunc("POST /radarr", s.radarrHandler)
	http.HandleFunc("POST /sonarr", s.sonarrHandler)
	return http.ListenAndServe(":"+s.listenAddr, nil)
}
