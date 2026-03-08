package parse

import (
	"fmt"
	"strings"

	"github.com/DeKoniX/telenews/models"
	"github.com/nicklaw5/helix/v2"
)

func (ParseNews *ParseNewsStruct) InitTwitch(clientID, clientSecret string) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return
	}

	resp, err := client.RequestAppAccessToken([]string{})
	if err != nil {
		return
	}

	client.SetAppAccessToken(resp.Data.AccessToken)
	ParseNews.twitch.client = client
}

func (ParseNews ParseNewsStruct) ParseTwitch(username string) ([]NewsStruct, error) {
	// Check if client is initialized
	if ParseNews.twitch.client == nil {
		return nil, fmt.Errorf("Twitch client not initialized")
	}

	// Call GetStreams to check if user is streaming
	resp, err := ParseNews.twitch.client.GetStreams(&helix.StreamsParams{
		UserLogins: []string{username},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get streams: %w", err)
	}

	// If no streams found, user is offline
	if len(resp.Data.Streams) == 0 {
		return []NewsStruct{}, nil
	}

	// Get the stream data
	stream := resp.Data.Streams[0]

	// Load source to check LastStreamID for deduplication
	var source models.Source
	err = source.SelectById(ParseNews.SourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source: %w", err)
	}

	// Check if this is the same stream (deduplication)
	if stream.ID == source.LastStreamID {
		return []NewsStruct{}, nil
	}

	// New stream detected - update LastStreamID
	source.LastStreamID = stream.ID
	err = source.Save()
	if err != nil {
		return nil, fmt.Errorf("failed to save LastStreamID: %w", err)
	}

	// Replace thumbnail URL template with actual dimensions
	thumbnailURL := strings.ReplaceAll(stream.ThumbnailURL, "{width}", "1920")
	thumbnailURL = strings.ReplaceAll(thumbnailURL, "{height}", "1080")

	// Create NewsStruct for the new stream
	news := NewsStruct{
		Title:    stream.Title,
		MSG:      fmt.Sprintf("Playing %s | %d viewers", stream.GameName, stream.ViewerCount),
		Link:     fmt.Sprintf("https://twitch.tv/%s", stream.UserLogin),
		PhotoURL: []string{thumbnailURL},
	}

	return []NewsStruct{news}, nil
}
