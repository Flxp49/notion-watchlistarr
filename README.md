# notion-watchlist-radarr-sonarr
Download media (via Radarr & Sonarr) directly from your watchlist on Notion.

# Prerequisites
- The script requires the following properties to exist in your watchlist database on Notion
  | Property Name (CASE SENSITIVE) | Property Type | Value |
  | -------- | -------- | -------- | 
  | `IMDb ID` | Text | IMDb id of series/movie |
  | `Type` | Select | `TV Series` or `Movie` |
  
- Radarr Webhook (**To be setup after running the app**)
  - Navigate to Connect under Radarr Settings
  - Add a new connection and choose Webhook
  - Subscribe to the following:
    - `On Grab`
    - `On Import`
    - `On Movie Added`
    - `On Movie Delete`
  - Set Webhook URL as `http://localhost:PORT/radarr` where `PORT` is whatever is set by you.
  - Set Method as `POST`
  - Click Save

- Sonarr Webhook (**To be setup after running the app**)
  - Navigate to Connect under Sonarr Settings
  - Add a new connection and choose Webhook
  - Subscribe to the following:
    - `On Grab`
    - `On Import`
    - `On Series Add`
    - `On Series Delete`
  - Set Webhook URL as `http://localhost:PORT/sonarr` where `PORT` is whatever is set by you.
  - Set Method as `POST`
  - Click Save

# Configuration
- Executable
- Docker

### Executable
The executable requires a .env file in the same directory from which it reads the following values:
| Env | Value | Default |
| -------- | -------- | -------- | 
| `PORT` | Port No to listen on | `7879` |
| `LOG_LEVEL` | `0` - DEBUG `1` - ERROR | `1` |
| `NOTION_INTEGRATION_SECRET` | Notion Integration Secret | NA |
| `NOTION_DB_ID` | Database id (found in the URL of the database page) | NA |
| `RADARR_HOST` | Radarr Host. Ex: `http://localhost:7878` | NA |
| `RADARR_KEY` | Radarr API key | NA |
| `RADARR_DEFAULT_ROOT_PATH` | Ex: `D:/Media/Movies` | If not provided, will set the first root path fetched from Radarr as default |
| `RADARR_DEFAULT_MONITOR` | Movie monitor profile, possible values: `MovieOnly` `MovieandCollection` | `MovieOnly` |
| `RADARR_DEFAULT_QUALITY_PROFILE`| Ex: `HD-1080p` | If not provided, will set the first profile fetched from Radarr as default |
| `SONARR_HOST` | Sonarr Host. Ex: `http://localhost:8989` | NA |
| `SONARR_KEY` | Sonarr API key | NA |
| `SONARR_DEFAULT_ROOT_PATH` | Ex: `D:/Media/Shows` | If not provided, will set the first root path fetched from Sonarr as default |
| `SONARR_DEFAULT_MONITOR` | TV monitor profile, possible values: `AllEpisodes` `FutureEpisodes` `MissingEpisodes` `ExistingEpisodes` `RecentEpisodes` `PilotEpisode` `FirstSeason` `LastSeason` `MonitorSpecials` `UnmonitorSpecials` `None` | `AllEpisodes` |
| `SONARR_DEFAULT_QUALITY_PROFILE` | Ex: `HD-1080p` | If not provided, will set the first profile fetched from Sonarr as default |
| `ARRSYNC_INTERVAL_SEC` | Duration (**Seconds**) Interval between each query to database for downloading | 10 |
| `WATCHLIST_SYNC_INTERVAL_HOUR` | Duration (**Hours**) Interval to sync media in Radarr and Sonarr library with watchlist | 12 |

### Docker
Env variables can be setup while spinning up the docker image. You can either set them via CLI individually or pass an env file.  
Ex: `docker run --env-file D:/path/to/env-file/.env -d -p 7879:7879 notionwatchlistarrsync:v0.1`  
env file is the same as the one specified above (Executable Section) EXCLUDING `PORT`
>**NOTE** the host for radarr and sonarr may have to be `http://host.docker.internal:XXXX` instead of `http://localhost:XXXX`

# Usage
The app on launch adds the following properties with values to the Notion database.

| Property Name | Property Type |
| -------- | -------- |  
| `Download` | Checkbox | 
| `Download Status` | Select | 
| `Quality Profile` | Select | 
| `Root Folder` | Select | 
| `Monitor` | Select | 

- `Quality Profile` is populated with the quality profiles fetched from Radarr and Sonarr as options.  
- `Root Folder` is populated with the root paths fetched from Radarr and Sonarr as options.  
- `Monitor` is populated with the following options:  
`TV Series: All Episodes`  
`TV Series: Future Episodes`  
`TV Series: Missing Episodes`  
`TV Series: Existing Episodes`  
`TV Series: Recent Episodes`  
`TV Series: PilotEpisode`  
`TV Series: FirstSeason`  
`TV Series: Last Season`  
`TV Series: Monitor Specials`  
`TV Series: Unmonitor Specials`  
`TV Series: None`  
`Movie: Movie Only`  
`Movie: Collection`

The options for the respective properties can be used to set the Quality Profile, Root Folder and Monitor Profile for the media to download (via Radarr/Sonarr) 

### Download
To download, 'Check' the Download property of the title.  
<img width="341" alt="image" src="https://github.com/Flxp49/notion-watchlist-radarr-sonarr/assets/63506727/18111ce8-5930-4add-998a-9d4f500ad0cb">  
Here, if `Quality Profile` `Root Folder` `Monitor` are not chosen before checking `Download`, the default values will be used instead.  
The app uses webhooks to sync the status of the media.

### Sync
- Sync: Queries the watchlist every `ARRSYNC_INTERVAL_SEC` for downloading media via Radarr/Sonarr
- LibrarySync: Syncs the existing media with the watchlist every `WATCHLIST_SYNC_INTERVAL_HOUR`
