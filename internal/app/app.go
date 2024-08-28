package app

import (
	"log/slog"
	"time"

	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/radarr"
	"github.com/flxp49/notion-watchlistarr/internal/sonarr"
)

type App struct {
	RadarrMedia  *RadarrMedia
	SonarrMedia  *SonarrMedia
	Logger       *slog.Logger
	PollInterval time.Duration
	SyncInterval time.Duration
	RadarrInit   bool
	SonarrInit   bool
}

func NewApp(N *notion.NotionClient, R *radarr.RadarrClient, S *sonarr.SonarrClient, Logger *slog.Logger, PollInterval time.Duration, SyncInterval time.Duration, RadarrInit bool, SonarrInit bool) *App {
	return &App{
		RadarrMedia:  NewRadarrMedia(N, R),
		SonarrMedia:  NewSonarrMedia(N, S),
		Logger:       Logger,
		PollInterval: PollInterval,
		SyncInterval: SyncInterval,
		RadarrInit:   RadarrInit,
		SonarrInit:   SonarrInit,
	}
}

func (A *App) RunApp() {
	if A.RadarrInit {
		go A.RadarrPollDB()
		go A.RadarrSyncWatchlist()
	}
	if A.SonarrInit {
		go A.SonarrPollDB()
		go A.SonarrSyncWatchlist()
	}
}

// Polls DB for titles from watchlist to download
func (A *App) RadarrPollDB() {
	for {
		A.Logger.Info("RadarrPollDB", "Status", "Fetching titles from database")
		notionPages, err := A.RadarrMedia.PollTitles()
		if err != nil {
			A.Logger.Error("RadarrPollDB", "Failed to query watchlist DB", err)
			time.Sleep(5 * time.Second)
			continue
		}
		A.Logger.Info("RadarrPollDB", "Status", "Fetched titles from DB", "No of titles fetched", len(notionPages.Results))
		for _, notionPage := range notionPages.Results {
			LookupData, LibraryData, err := A.RadarrMedia.ProcessTitles(notionPage)
			if err != nil {
				A.Logger.Error("RadarrPollDB", "Failed to process movie in Radarr", notionPage.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
				A.RadarrMedia.N.UpdateDownloadStatus("movie", notionPage.Pgid, false, "Error", "", "", "")
				continue
			}
			if len(LibraryData) != 0 {
				err = A.RadarrMedia.HandleExistingTitle(LibraryData, notionPage)
				if err != nil {
					A.Logger.Error("RadarrPollDB", "Failed to handle existing movie in Radarr", notionPage.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
					A.RadarrMedia.N.UpdateDownloadStatus("movie", notionPage.Pgid, false, "Error", "", "", "")
					continue
				}
			}
			err = A.RadarrMedia.AddTitle(LookupData, notionPage)
			if err != nil {
				A.Logger.Error("RadarrPollDB", "Failed to add movie to Radarr", notionPage.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
				A.RadarrMedia.N.UpdateDownloadStatus("movie", notionPage.Pgid, false, "Error", "", "", "")
			}
		}
		time.Sleep(A.PollInterval * time.Second)
	}
}
func (A *App) SonarrPollDB() {
	for {
		A.Logger.Info("SonarrPollDB", "Status", "Fetching titles from database")
		notionPages, err := A.SonarrMedia.PollTitles()
		if err != nil {
			A.Logger.Error("SonarrPollDB", "Failed to query watchlist DB", err)
			time.Sleep(5 * time.Second)
			continue
		}
		A.Logger.Info("SonarrPollDB", "Status", "Fetched titles from DB", "No of titles fetched", len(notionPages.Results))
		for _, notionPage := range notionPages.Results {
			LookupData, LibraryData, err := A.SonarrMedia.ProcessTitles(notionPage)
			if err != nil {
				A.Logger.Error("SonarrPollDB", "Failed to process movie in Sonarr", notionPage.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
				A.SonarrMedia.N.UpdateDownloadStatus("series", notionPage.Pgid, false, "Error", "", "", "")
				continue
			}
			if len(LibraryData) != 0 {
				err = A.SonarrMedia.HandleExistingTitle(LibraryData, notionPage)
				if err != nil {
					A.Logger.Error("SonarrPollDB", "Failed to handle existing movie in Sonarr", notionPage.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
					A.SonarrMedia.N.UpdateDownloadStatus("series", notionPage.Pgid, false, "Error", "", "", "")
					continue
				}
			}
			err = A.SonarrMedia.AddTitle(LookupData, notionPage)
			if err != nil {
				A.Logger.Error("SonarrPollDB", "Failed to add movie to Sonarr", notionPage.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
				A.SonarrMedia.N.UpdateDownloadStatus("series", notionPage.Pgid, false, "Error", "", "", "")
			}
		}
		time.Sleep(A.PollInterval * time.Second)
	}
}

// Sync Radarr library with watchlist
func (A *App) RadarrSyncWatchlist() {
	for {
		A.Logger.Info("RadarrSyncWatchlist", "Status", "Fetching titles from Radarr")
		radarrLibrary, err := A.RadarrMedia.FetchRadarrLibrary()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		A.Logger.Info("RadarrSyncWatchlist", "Status", "Fetched titles from DB", "No of titles fetched", len(radarrLibrary))
		for _, radarrMovie := range radarrLibrary {
			watchlistMovie, err := A.RadarrMedia.N.QueryDBImdb(radarrMovie.ImdbID)
			if err != nil {
				A.Logger.Error("RadarrSyncWatchlist", "Failed to query movie from notion watchlist", err)
				continue
			}
			if len(watchlistMovie.Results) == 0 {
				continue
			}
			err = A.RadarrMedia.ProcessLibraryTitle(watchlistMovie, radarrMovie)
			if err != nil {
				A.Logger.Error("RadarrSyncWatchlist", "Failed to process movie", err)
				continue
			}
		}
		A.Logger.Info("RadarrSyncWatchlist", "Status", "Finished")
		time.Sleep(A.SyncInterval * time.Hour)
	}
}

func (A *App) SonarrSyncWatchlist() {
	for {
		A.Logger.Info("SonarrSyncWatchlist", "Status", "Fetching titles from Sonarr")
		sonarrLibrary, err := A.SonarrMedia.FetchSonarrLibrary()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		A.Logger.Info("SonarrSyncWatchlist", "Status", "Fetched titles from DB", "No of titles fetched", len(sonarrLibrary))
		for _, sonarrSeries := range sonarrLibrary {
			watchlistSeries, err := A.SonarrMedia.N.QueryDBImdb(sonarrSeries.ImdbID)
			if err != nil {
				A.Logger.Error("SonarrSyncWatchlist", "Failed to query series from notion watchlist", err)
				continue
			}
			if len(watchlistSeries.Results) == 0 {
				continue
			}
			err = A.SonarrMedia.ProcessLibraryTitle(watchlistSeries, sonarrSeries)
			if err != nil {
				A.Logger.Error("SonarrSyncWatchlist", "Failed to process series", err)
				continue
			}
		}
		A.Logger.Info("SonarrSyncWatchlist", "Status", "Finished")
		time.Sleep(A.SyncInterval * time.Hour)
	}
}
