package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// AudioMetadata matches the TypeScript interface
type AudioMetadata struct {
	Language        string   `json:"language"`
	SpeakerGender   string   `json:"speakerGender,omitempty"`
	SpeakerAgeRange string   `json:"speakerAgeRange,omitempty"`
	SpeakerDialect  string   `json:"speakerDialect,omitempty"`
	Transcript      string   `json:"transcript,omitempty"`
	SourceUrl       string   `json:"sourceUrl,omitempty"`
	Tags            []string `json:"tags"`
}

// AudioClip matches the TypeScript interface
type AudioClip struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Duration         float64         `json:"duration"`
	Filename         string          `json:"filename"`
	OriginalFilename string          `json:"originalFilename"`
	FileSize         int64           `json:"fileSize"`
	Metadata         AudioMetadata   `json:"metadata"`
	UploadedBy       string          `json:"uploadedBy"`
	CreatedAt        string          `json:"createdAt"`
	UpdatedAt        string          `json:"updatedAt"`
	URL              string          `json:"url,omitempty"` // Added by API
}

// API response structure
type ClipsResponse struct {
	Clips []AudioClip `json:"clips"`
}

// Scraper configuration
type Scraper struct {
	baseURL      string
	outputDir    string
	audioDir     string
	language     string
	limit        int
	batchSize    int
	client       *http.Client
	downloadAudio bool
}

// NewScraper creates a new scraper instance
func NewScraper(baseURL, outputDir string, downloadAudio bool) *Scraper {
	return &Scraper{
		baseURL:       baseURL,
		outputDir:     outputDir,
		audioDir:      filepath.Join(outputDir, "audio"),
		client:        &http.Client{Timeout: 30 * time.Second},
		downloadAudio: downloadAudio,
	}
}

// Setup creates necessary directories
func (s *Scraper) Setup() error {
	if err := os.MkdirAll(s.audioDir, 0755); err != nil {
		return fmt.Errorf("failed to create audio directory: %v", err)
	}
	return nil
}

// FetchClips fetches clips from the API efficiently with a large limit
func (s *Scraper) FetchClips(filters map[string]string, limit int) ([]AudioClip, error) {
	var allClips []AudioClip

	fmt.Printf("🔍 Fetching all clips from %s...\n", s.baseURL)

	params := url.Values{}
	// Use a very large limit to fetch all clips in one request (100000 should cover all)
	params.Set("limit", "100000")
	params.Set("offset", "0")

	// Add custom filters
	for key, value := range filters {
		if value != "" {
			params.Set(key, value)
		}
	}

	apiURL := fmt.Sprintf("%s/api/clips?%s", s.baseURL, params.Encode())
	fmt.Printf("  📥 Fetching all clips from: %s\n", apiURL)

	resp, err := s.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clips: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result ClipsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	allClips = append(allClips, result.Clips...)
	fmt.Printf("  ✅ Retrieved %d clips\n", len(allClips))

	if limit > 0 && len(allClips) > limit {
		allClips = allClips[:limit]
		fmt.Printf("  ⚠️  Limited to %d clips\n", limit)
	}

	return allClips, nil
}

// DownloadAudio downloads audio file for a clip
func (s *Scraper) DownloadAudio(clip AudioClip) (string, error) {
	if !s.downloadAudio {
		return "", nil
	}

	fileURL := fmt.Sprintf("%s%s", s.baseURL, clip.URL)
	filename := filepath.Join(s.audioDir, clip.Filename)

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("    ⏭️  Skipping (already exists): %s\n", clip.Filename)
		return filename, nil
	}

	fmt.Printf("    📥 Downloading: %s\n", clip.Filename)

	resp, err := s.client.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %v", clip.Filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d for %s", resp.StatusCode, clip.Filename)
	}

	out, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write file %s: %v", filename, err)
	}

	return filename, nil
}

// SaveClipsJSON saves clips metadata as JSON
func (s *Scraper) SaveClipsJSON(clips []AudioClip) (string, error) {
	filename := filepath.Join(s.outputDir, "clips.json")

	data, err := json.MarshalIndent(clips, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON file: %v", err)
	}

	fmt.Printf("✅ Saved clips metadata: %s\n", filename)
	return filename, nil
}

// SaveClipsCSV saves clips as CSV for easy viewing
func (s *Scraper) SaveClipsCSV(clips []AudioClip) (string, error) {
	filename := filepath.Join(s.outputDir, "clips.csv")

	out, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer out.Close()

	// Write header
	header := "ID,Title,Duration(s),Language,Filename,UploadedBy,CreatedAt\n"
	if _, err := out.WriteString(header); err != nil {
		return "", err
	}

	// Write rows
	for _, clip := range clips {
		row := fmt.Sprintf("\"%s\",\"%s\",%.2f,\"%s\",\"%s\",\"%s\",\"%s\"\n",
			clip.ID, clip.Title, clip.Duration, clip.Metadata.Language,
			clip.Filename, clip.UploadedBy, clip.CreatedAt)
		if _, err := out.WriteString(row); err != nil {
			return "", err
		}
	}

	fmt.Printf("✅ Saved clips as CSV: %s\n", filename)
	return filename, nil
}

// PrintStatistics prints scraping statistics
func (s *Scraper) PrintStatistics(clips []AudioClip) {
	if len(clips) == 0 {
		return
	}

	languageMap := make(map[string]int)
	genderMap := make(map[string]int)
	totalDuration := 0.0

	for _, clip := range clips {
		languageMap[clip.Metadata.Language]++
		if clip.Metadata.SpeakerGender != "" {
			genderMap[clip.Metadata.SpeakerGender]++
		}
		totalDuration += clip.Duration
	}

	fmt.Println("\n📊 Scraping Statistics:")
	fmt.Printf("  Total clips: %d\n", len(clips))
	fmt.Printf("  Total duration: %.2f minutes (%.2f hours)\n", totalDuration/60, totalDuration/3600)
	fmt.Println("  Languages:")
	for lang, count := range languageMap {
		fmt.Printf("    - %s: %d clips\n", lang, count)
	}
	if len(genderMap) > 0 {
		fmt.Println("  Speaker genders:")
		for gender, count := range genderMap {
			fmt.Printf("    - %s: %d clips\n", gender, count)
		}
	}
}

// Run performs the complete scraping operation
func (s *Scraper) Run(filters map[string]string, limit int) error {
	if err := s.Setup(); err != nil {
		return err
	}

	// Fetch clips
	clips, err := s.FetchClips(filters, limit)
	if err != nil {
		return err
	}

	if len(clips) == 0 {
		fmt.Println("⚠️  No clips found!")
		return nil
	}

	fmt.Printf("\n🎯 Processing %d clips...\n", len(clips))

	// Download audio files
	if s.downloadAudio {
		for i, clip := range clips {
			fmt.Printf("[%d/%d] ", i+1, len(clips))
			if _, err := s.DownloadAudio(clip); err != nil {
				fmt.Printf("⚠️  %v\n", err)
				// Continue with other clips even if one fails
			} else {
				fmt.Printf("✅ Downloaded: %s\n", clip.Title)
			}
		}
	}

	// Save metadata
	if _, err := s.SaveClipsJSON(clips); err != nil {
		return err
	}

	if _, err := s.SaveClipsCSV(clips); err != nil {
		return err
	}

	// Print statistics
	s.PrintStatistics(clips)

	return nil
}

func main() {
	baseURL := flag.String("url", "https://the-chorusing-lab.vercel.app", "Base URL of the chorusing lab site")
	outputDir := flag.String("output", "./scraped_clips", "Output directory for scraped clips")
	language := flag.String("language", "", "Filter by language (optional)")
	limit := flag.Int("limit", 0, "Maximum number of clips to scrape (0 = unlimited)")
	downloadAudio := flag.Bool("download", false, "Download audio files (default: metadata only)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Chorus Lab Scraper - Download clips from a Chorus Lab instance\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Get metadata for all clips\n")
		fmt.Fprintf(os.Stderr, "  %s -url https://the-chorusing-lab.vercel.app\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Download English clips with audio\n")
		fmt.Fprintf(os.Stderr, "  %s -language English -download\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Limit to 100 clips and save to custom directory\n")
		fmt.Fprintf(os.Stderr, "  %s -limit 100 -output ./my_clips -download\n", os.Args[0])
	}

	flag.Parse()

	fmt.Println("🎵 Chorus Lab Scraper")
	fmt.Println("====================")

	scraper := NewScraper(*baseURL, *outputDir, *downloadAudio)

	filters := make(map[string]string)
	if *language != "" {
		filters["language"] = *language
	}

	if err := scraper.Run(filters, *limit); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✨ Scraping complete!")
	fmt.Printf("📁 Output saved to: %s\n", *outputDir)
}
