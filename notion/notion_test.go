package notion

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var Notion *NotionClient

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Notion = InitNotionClient(os.Getenv("NOTION_USER"), os.Getenv("NOTION_INTEGRATION_SECRET"), os.Getenv("NOTION_DB_ID"))
	os.Exit(m.Run())
}
