package upload

import (
	"fmt"
	"io"
	"time"

	"github.com/schollz/progressbar/v3"

	"cfstream/internal/api"
)

// ProgressTracker wraps a progress bar and handles upload progress updates.
type ProgressTracker struct {
	bar       *progressbar.ProgressBar
	startTime time.Time
	quiet     bool
}

// NewProgressTracker creates a new progress tracker for file uploads.
func NewProgressTracker(fileSize int64, filename string, quiet bool) *ProgressTracker {
	if quiet {
		return &ProgressTracker{
			quiet:     true,
			startTime: time.Now(),
		}
	}

	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription(fmt.Sprintf("Uploading %s", filename)),
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	return &ProgressTracker{
		bar:       bar,
		startTime: time.Now(),
		quiet:     quiet,
	}
}

// Update updates the progress bar with the current upload progress.
func (pt *ProgressTracker) Update(progress api.UploadProgress) {
	if pt.quiet {
		return
	}

	if pt.bar != nil {
		_ = pt.bar.Set64(progress.BytesSent) //nolint:errcheck // Progress bar errors are not critical
	}
}

// Finish marks the upload as complete.
func (pt *ProgressTracker) Finish() {
	if pt.quiet {
		return
	}

	if pt.bar != nil {
		_ = pt.bar.Finish() //nolint:errcheck // Progress bar errors are not critical
	}
}

// Duration returns the time elapsed since the tracker was created.
func (pt *ProgressTracker) Duration() time.Duration {
	return time.Since(pt.startTime)
}

// FormatBytes formats a byte count in human-readable format.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSpeed formats upload speed in human-readable format.
func FormatSpeed(bytesPerSecond float64) string {
	return fmt.Sprintf("%s/s", FormatBytes(int64(bytesPerSecond)))
}
