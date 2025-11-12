package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/stream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Client interface for testing.
type MockClient struct {
	mock.Mock
}

func (m *MockClient) ListVideos(ctx context.Context, opts *ListOptions) ([]Video, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Video), args.Error(1)
}

func (m *MockClient) GetVideo(ctx context.Context, videoID string) (*Video, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Video), args.Error(1)
}

func (m *MockClient) DeleteVideo(ctx context.Context, videoID string) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

// Test NewClient validation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		apiToken  string
		wantErr   bool
	}{
		{
			name:      "valid credentials",
			accountID: "test-account-id",
			apiToken:  "test-api-token",
			wantErr:   false,
		},
		{
			name:      "missing account ID",
			accountID: "",
			apiToken:  "test-api-token",
			wantErr:   true,
		},
		{
			name:      "missing API token",
			accountID: "test-account-id",
			apiToken:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.accountID, tt.apiToken)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// Test VideoFromSDK conversion
func TestVideoFromSDK(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    *stream.Video
		expected *Video
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "basic video with name in meta",
			input: &stream.Video{
				UID:           "test-uid-123",
				Duration:      120.5,
				Created:       now,
				Modified:      now,
				ReadyToStream: true,
				Preview:       "https://example.com/preview",
				Thumbnail:     "https://example.com/thumbnail",
				Status: stream.VideoStatus{
					State: stream.VideoStatusStateReady,
				},
				Meta: map[string]interface{}{
					"name": "Test Video",
				},
			},
			expected: &Video{
				UID:           "test-uid-123",
				Name:          "Test Video",
				Status:        "ready",
				Duration:      120.5,
				Created:       now,
				Modified:      now,
				ReadyToStream: true,
				Preview:       "https://example.com/preview",
				Thumbnail:     "https://example.com/thumbnail",
				Meta: map[string]interface{}{
					"name": "Test Video",
				},
			},
		},
		{
			name: "video without name uses UID",
			input: &stream.Video{
				UID:      "test-uid-456",
				Duration: 60.0,
				Created:  now,
				Modified: now,
				Status: stream.VideoStatus{
					State: stream.VideoStatusStateQueued,
				},
			},
			expected: &Video{
				UID:      "test-uid-456",
				Name:     "test-uid-456",
				Status:   "queued",
				Duration: 60.0,
				Created:  now,
				Modified: now,
			},
		},
		{
			name: "video with error status",
			input: &stream.Video{
				UID:      "test-uid-789",
				Duration: 0,
				Created:  now,
				Modified: now,
				Status: stream.VideoStatus{
					State:           stream.VideoStatusStateError,
					ErrorReasonText: "encoding failed",
				},
			},
			expected: &Video{
				UID:           "test-uid-789",
				Name:          "test-uid-789",
				Status:        "error",
				StatusDetails: "encoding failed",
				Duration:      0,
				Created:       now,
				Modified:      now,
			},
		},
		{
			name: "video in progress",
			input: &stream.Video{
				UID:      "test-uid-999",
				Duration: 0,
				Created:  now,
				Modified: now,
				Status: stream.VideoStatus{
					State:       stream.VideoStatusStateInprogress,
					PctComplete: "45",
				},
			},
			expected: &Video{
				UID:           "test-uid-999",
				Name:          "test-uid-999",
				Status:        "inprogress",
				StatusDetails: "45% complete",
				Duration:      0,
				Created:       now,
				Modified:      now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VideoFromSDK(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test VideosFromSDK conversion
func TestVideosFromSDK(t *testing.T) {
	now := time.Now()

	input := []stream.Video{
		{
			UID:      "test-1",
			Duration: 100,
			Created:  now,
			Modified: now,
			Status: stream.VideoStatus{
				State: stream.VideoStatusStateReady,
			},
		},
		{
			UID:      "test-2",
			Duration: 200,
			Created:  now,
			Modified: now,
			Status: stream.VideoStatus{
				State: stream.VideoStatusStateQueued,
			},
		},
	}

	result := VideosFromSDK(input)
	assert.Len(t, result, 2)
	assert.Equal(t, "test-1", result[0].UID)
	assert.Equal(t, "test-2", result[1].UID)
}

// Test WrapError function
func TestWrapError(t *testing.T) {
	tests := []struct {
		name            string
		input           error
		expectedErr     error
		checkErrorChain bool
	}{
		{
			name:            "nil error",
			input:           nil,
			expectedErr:     nil,
			checkErrorChain: false,
		},
		{
			name:            "non-Cloudflare error",
			input:           errors.New("generic error"),
			expectedErr:     nil,
			checkErrorChain: false,
		},
		{
			name: "404 not found",
			input: &cloudflare.Error{
				StatusCode: http.StatusNotFound,
			},
			expectedErr:     ErrNotFound,
			checkErrorChain: true,
		},
		{
			name: "401 unauthorized",
			input: &cloudflare.Error{
				StatusCode: http.StatusUnauthorized,
			},
			expectedErr:     ErrUnauthorized,
			checkErrorChain: true,
		},
		{
			name: "403 forbidden",
			input: &cloudflare.Error{
				StatusCode: http.StatusForbidden,
			},
			expectedErr:     ErrForbidden,
			checkErrorChain: true,
		},
		{
			name: "429 rate limit",
			input: &cloudflare.Error{
				StatusCode: http.StatusTooManyRequests,
			},
			expectedErr:     ErrRateLimit,
			checkErrorChain: true,
		},
		{
			name: "400 bad request",
			input: &cloudflare.Error{
				StatusCode: http.StatusBadRequest,
			},
			expectedErr:     ErrInvalidInput,
			checkErrorChain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.input)
			if tt.expectedErr == nil && tt.input == nil {
				assert.NoError(t, result)
			} else if tt.checkErrorChain {
				assert.Error(t, result)
				assert.ErrorIs(t, result, tt.expectedErr)
			} else {
				// For non-Cloudflare errors, they should pass through unchanged
				assert.Equal(t, tt.input, result)
			}
		})
	}
}

// Test MockClient usage
func TestMockClient(t *testing.T) {
	ctx := context.Background()

	t.Run("mock ListVideos", func(t *testing.T) {
		mockClient := new(MockClient)
		expectedVideos := []Video{
			{UID: "video-1", Name: "Test Video 1"},
			{UID: "video-2", Name: "Test Video 2"},
		}

		mockClient.On("ListVideos", ctx, (*ListOptions)(nil)).Return(expectedVideos, nil)

		videos, err := mockClient.ListVideos(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedVideos, videos)
		mockClient.AssertExpectations(t)
	})

	t.Run("mock GetVideo", func(t *testing.T) {
		mockClient := new(MockClient)
		expectedVideo := &Video{UID: "video-1", Name: "Test Video"}

		mockClient.On("GetVideo", ctx, "video-1").Return(expectedVideo, nil)

		video, err := mockClient.GetVideo(ctx, "video-1")
		assert.NoError(t, err)
		assert.Equal(t, expectedVideo, video)
		mockClient.AssertExpectations(t)
	})

	t.Run("mock DeleteVideo", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("DeleteVideo", ctx, "video-1").Return(nil)

		err := mockClient.DeleteVideo(ctx, "video-1")
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("mock error handling", func(t *testing.T) {
		mockClient := new(MockClient)
		expectedErr := ErrNotFound

		mockClient.On("GetVideo", ctx, "nonexistent").Return((*Video)(nil), expectedErr)

		video, err := mockClient.GetVideo(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, video)
		assert.ErrorIs(t, err, ErrNotFound)
		mockClient.AssertExpectations(t)
	})
}
