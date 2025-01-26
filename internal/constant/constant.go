package constant

const (
	// SONARR Monitor Profile Constants
	AllEpisodes       = "All"
	FutureEpisodes    = "Future"
	MissingEpisodes   = "Missing"
	ExistingEpisodes  = "Existing"
	RecentEpisodes    = "Recent"
	PilotEpisode      = "Pilot"
	FirstSeason       = "FirstSeason"
	LastSeason        = "LastSeason"
	MonitorSpecials   = "MonitorSpecials"
	UnmonitorSpecials = "UnmonitorSpecials"
	// RADARR Monitor Profile Constants
	MovieOnly          = "MovieOnly"
	MovieAndCollection = "MovieandCollection"
	// Notion DB Select Option Names
	NotionOptionAllEpisodes       = "TV Series: All Episodes"
	NotionOptionFutureEpisodes    = "TV Series: Future Episodes"
	NotionOptionMissingEpisodes   = "TV Series: Missing Episodes"
	NotionOptionExistingEpisodes  = "TV Series: Existing Episodes"
	NotionOptionRecentpisodes     = "TV Series: Recent Episodes"
	NotionOptionPilotEpisode      = "TV Series: Pilot Episode"
	NotionOptionFirstSeason       = "TV Series: First Season"
	NotionOptionLastSeason        = "TV Series: Last Season"
	NotionOptionMonitorSpecials   = "TV Series: Monitor Specials"
	NotionOptionUnmonitorSpecials = "TV Series: Unmonitor Specials"
	NotionOptionMovieOnly         = "Movie: Movie Only"
	NotionOptionCollection        = "Movie: Collection"

	MediaTypeMovie           = "Movie"
	MediaTypeTV              = "TV Series"
	MediaStatusDownloaded    = "Downloaded"
	MediaStatusDownloading   = "Downloading"
	MediaStatusNotDownloaded = "Not Downloaded"
	MediaStatusQueued        = "Queued"
	MediaStatusError         = "Error"

	EventTypeTest            = "Test"
	EventTypeMovieAdded      = "MovieAdded"
	EventTypeMovieGrabbed    = "Grab"
	EventTypeMovieDownloaded = "Download"
	EventTypeMovieDelete     = "MovieDelete"
	EventTypeMovieFileDelete = "MovieFileDelete"
	EventTypeTVAdded         = "SeriesAdd"
	EventTypeTVGrabbed       = "Grab"
	EventTypeTVDownloaded    = "Download"
	EventTypeTVDelete        = "SeriesDelete"

	IMDB = "imdb"
	TVDB = "tvdb"
)
