package parse

import (
	"os"
	"testing"

	"github.com/DeKoniX/telenews/models"
)

func TestInitTwitch(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		description  string
	}{
		{
			name:         "Empty credentials",
			clientID:     "",
			clientSecret: "",
			description:  "Should fail at NewClient or token request",
		},
		{
			name:         "Non-empty credentials",
			clientID:     "test_client_id",
			clientSecret: "test_secret",
			description:  "May succeed at init but fail on API calls",
		},
		{
			name:         "Valid credentials from env",
			clientID:     os.Getenv("TwitchClientID"),
			clientSecret: os.Getenv("TwitchClientSecret"),
			description:  "Should succeed if env vars are valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parseNews ParseNewsStruct
			parseNews.InitTwitch(tt.clientID, tt.clientSecret)

			t.Logf("%s: client initialized = %v", tt.description, parseNews.twitch.client != nil)
		})
	}
}

func TestParseTwitch_NilClient(t *testing.T) {
	var parseNews ParseNewsStruct

	news, err := parseNews.ParseTwitch("test_user")

	if err == nil {
		t.Error("Expected error when client is not initialized")
	}
	if news != nil {
		t.Errorf("Expected nil news slice, got %v", news)
	}
	if err.Error() != "Twitch client not initialized" {
		t.Errorf("Expected 'Twitch client not initialized' error, got: %v", err)
	}
}

func TestParseTwitch_Scenarios(t *testing.T) {
	t.Run("Nil client returns error", func(t *testing.T) {
		var parseNews ParseNewsStruct
		_, err := parseNews.ParseTwitch("any_user")
		if err == nil || err.Error() != "Twitch client not initialized" {
			t.Errorf("Expected 'Twitch client not initialized', got: %v", err)
		}
	})

	t.Run("Empty username with nil client", func(t *testing.T) {
		var parseNews ParseNewsStruct
		_, err := parseNews.ParseTwitch("")
		if err == nil || err.Error() != "Twitch client not initialized" {
			t.Errorf("Expected 'Twitch client not initialized', got: %v", err)
		}
	})
}

// TestParseTwitch_Integration tests ParseTwitch with real API
// Requires: TwitchClientID, TwitchClientSecret env vars
// Requires: Database connection (DBADDRESS, DBNAME, DBUSER, DBPASS)
func TestParseTwitch_Integration(t *testing.T) {
	clientID := os.Getenv("TwitchClientID")
	clientSecret := os.Getenv("TwitchClientSecret")

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: TwitchClientID or TwitchClientSecret not set")
	}

	// Initialize database
	err := connectDB()
	if err != nil {
		t.Skip("Skipping integration test: Database not available")
	}

	// Create test user and source
	user, err := createTestUser("TwitchTestUser", 999999)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer user.Delete()

	source, err := createTestSource(user, "test_streamer", models.Twitch)
	if err != nil {
		t.Fatalf("Failed to create test source: %v", err)
	}
	defer source.Delete()

	// Initialize parser
	var parseNews ParseNewsStruct
	parseNews.SourceID = source.ID
	parseNews.InitTwitch(clientID, clientSecret)

	if parseNews.twitch.client == nil {
		t.Fatal("Failed to initialize Twitch client with valid credentials")
	}

	// Test with a known streamer (may be offline)
	testUsername := "shroud" // Popular streamer, likely to exist

	news, err := parseNews.ParseTwitch(testUsername)
	if err != nil {
		t.Errorf("ParseTwitch failed: %v", err)
	}

	// Result can be empty (offline) or contain one stream (online)
	if len(news) > 1 {
		t.Errorf("Expected 0 or 1 news items, got %d", len(news))
	}

	if len(news) == 1 {
		// Validate NewsStruct fields
		if news[0].Title == "" {
			t.Error("Expected non-empty Title")
		}
		if news[0].MSG == "" {
			t.Error("Expected non-empty MSG")
		}
		if news[0].Link == "" {
			t.Error("Expected non-empty Link")
		}
		if len(news[0].PhotoURL) == 0 {
			t.Error("Expected at least one PhotoURL")
		}

		// Test deduplication - calling again should return empty
		news2, err := parseNews.ParseTwitch(testUsername)
		if err != nil {
			t.Errorf("Second ParseTwitch call failed: %v", err)
		}
		if len(news2) != 0 {
			t.Error("Expected empty result on duplicate stream (same Stream ID)")
		}
	}
}

// TestParseTwitch_InvalidUsername tests ParseTwitch with non-existent user
func TestParseTwitch_InvalidUsername(t *testing.T) {
	clientID := os.Getenv("TwitchClientID")
	clientSecret := os.Getenv("TwitchClientSecret")

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping test: TwitchClientID or TwitchClientSecret not set")
	}

	// Initialize database
	err := connectDB()
	if err != nil {
		t.Skip("Skipping test: Database not available")
	}

	// Create test user and source
	user, err := createTestUser("TwitchTestUser2", 999998)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer user.Delete()

	source, err := createTestSource(user, "nonexistent_user_12345", models.Twitch)
	if err != nil {
		t.Fatalf("Failed to create test source: %v", err)
	}
	defer source.Delete()

	// Initialize parser
	var parseNews ParseNewsStruct
	parseNews.SourceID = source.ID
	parseNews.InitTwitch(clientID, clientSecret)

	if parseNews.twitch.client == nil {
		t.Fatal("Failed to initialize Twitch client")
	}

	// Test with non-existent username
	news, err := parseNews.ParseTwitch("this_user_definitely_does_not_exist_12345")

	// Should not error - just return empty (offline)
	if err != nil {
		t.Errorf("Expected no error for non-existent user, got: %v", err)
	}
	if len(news) != 0 {
		t.Errorf("Expected empty news for non-existent user, got %d items", len(news))
	}
}

// Helper functions for database setup

func connectDB() error {
	var envDBAddress = os.Getenv("DBADDRESS")
	var envDBName = os.Getenv("DBNAME")
	var envDBUser = os.Getenv("DBUSER")
	var envDBPass = os.Getenv("DBPASS")

	if envDBAddress == "" || envDBName == "" {
		return nil // Skip if DB not configured
	}

	err := models.Init(envDBAddress, envDBUser, envDBPass, envDBName)
	return err
}

func createTestUser(username string, chatID int64) (models.User, error) {
	user := models.User{UserName: username, ChatID: chatID}
	err := user.Insert()
	return user, err
}

func createTestSource(user models.User, query string, sourceType models.SourceType) (models.Source, error) {
	source := models.Source{Query: query, Type: sourceType}
	_, err := source.Insert(user)
	return source, err
}
