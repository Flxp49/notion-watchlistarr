package api

import (
	"log/slog"
	"net/http"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
)

type Server struct {
	listenAddr string
	N          *notion.NotionClient
	R          *radarr.RadarrClient
	Logger     *slog.Logger
}

func NewServer(listenAddr string, N *notion.NotionClient, R *radarr.RadarrClient, Logger *slog.Logger) *Server {
	return &Server{
		listenAddr: listenAddr,
		N:          N,
		R:          R,
		Logger:     Logger,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("POST /radarr", s.radarrHandler)
	// http.HandleFunc("POST /sonarr", sonarrHandler)
	http.HandleFunc("/", s.incorrectReqHandler)
	return http.ListenAndServe(":"+s.listenAddr, nil)
}
