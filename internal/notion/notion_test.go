package notion

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var Notion *NotionClient

func TestMain(m *testing.M) {
	err := godotenv.Load("../../cmd/notionwatchlistarrsync/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Notion = InitNotionClient(os.Getenv("NOTION_USER"), os.Getenv("NOTION_INTEGRATION_SECRET"), os.Getenv("NOTION_DB_ID"))
	os.Exit(m.Run())
}

func TestGetNotionMonitorProp(t *testing.T) {
	prop, err := Notion.GetNotionMonitorProp("MovieandCollection", "Movie")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(prop)
}

func TestQueryDB(t *testing.T) {
	// series, err := Notion.QueryDB("TV Series")
	series, err := Notion.QueryDB("Movie")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(series)
}

func TestQueryDBTmdb(t *testing.T) {
	series, err := Notion.QueryDBTmdb(213241)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(series)
}
func TestQueryDBImdb(t *testing.T) {
	series, err := Notion.QueryDBImdb("tt13802576")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(series)
}
