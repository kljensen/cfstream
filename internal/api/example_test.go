package api_test

import (
	"context"
	"fmt"

	"cfstream/internal/api"
)

// Example demonstrates basic usage of the Cloudflare Stream API client.
func Example() {
	// Create a new client with your credentials
	client, err := api.NewClient("your-account-id", "your-api-token")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// List all videos
	videos, err := client.ListVideos(ctx, nil)
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		fmt.Printf("Video: %s (%s)\n", video.Name, video.Status)
	}
}

// Example_listWithOptions demonstrates listing videos with search filters.
func Example_listWithOptions() {
	client, err := api.NewClient("your-account-id", "your-api-token")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// List videos with search filter
	videos, err := client.ListVideos(ctx, &api.ListOptions{
		Search: "tutorial",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d videos matching 'tutorial'\n", len(videos))
}

// Example_getVideo demonstrates fetching a single video by ID.
func Example_getVideo() {
	client, err := api.NewClient("your-account-id", "your-api-token")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Get a specific video
	video, err := client.GetVideo(ctx, "video-uid-123")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Video: %s\n", video.Name)
	fmt.Printf("Status: %s\n", video.Status)
	fmt.Printf("Duration: %.2f seconds\n", video.Duration)
}

// Example_deleteVideo demonstrates deleting a video.
func Example_deleteVideo() {
	client, err := api.NewClient("your-account-id", "your-api-token")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Delete a video
	err = client.DeleteVideo(ctx, "video-uid-123")
	if err != nil {
		panic(err)
	}

	fmt.Println("Video deleted successfully")
}
