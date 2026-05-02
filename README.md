# Chorus Lab Scraper

A Go command-line tool to scrape clips from a Chorus Lab instance (including the one at https://the-chorusing-lab.vercel.app).

## Features

- ✅ Fetch clip metadata from the API
- ✅ Download audio files (optional)
- ✅ Filter by language and other criteria
- ✅ Export as JSON and CSV
- ✅ Automatic pagination handling
- ✅ Rate limiting (respectful 500ms delays between requests)
- ✅ Detailed logging and statistics

## Installation

### Prerequisites

- Go 1.19 or higher

### Building

```bash
# Navigate to the project root
cd d:\Tech\Projects\the-chorusing-lab

# Build the scraper
go build -o scraper.exe scraper.go

# Or build on macOS/Linux
go build -o scraper scraper.go
```

## Usage

### Basic Usage - Get metadata only

```bash
# Get all clips (metadata only)
./scraper.exe

# Specify a custom URL
./scraper.exe -url https://the-chorusing-lab.vercel.app

# Filter by language
./scraper.exe -language English
```

### Download Audio

```bash
# Download clips with audio files
./scraper.exe -download

# Download limited number of clips
./scraper.exe -download -limit 50

# Download English clips to a custom directory
./scraper.exe -language English -download -output ./my_english_clips
```

### All Options

```bash
./scraper.exe -h
```

Options:
- `-url string` - Base URL of the chorusing lab site (default: `https://the-chorusing-lab.vercel.app`)
- `-output string` - Output directory for scraped clips (default: `./scraped_clips`)
- `-language string` - Filter by language (optional)
- `-limit int` - Maximum number of clips to scrape (0 = unlimited, default: `0`)
- `-download` - Download audio files (default: false, metadata only)

## Output

The scraper creates the following structure:

```
scraped_clips/
├── clips.json          # Clip metadata as JSON
├── clips.csv           # Clip metadata as CSV
└── audio/              # Downloaded audio files (if -download flag used)
    ├── clip-1.mp3
    ├── clip-2.mp3
    └── ...
```

### clips.json Example

```json
[
  {
    "id": "clip-1234567890-abc123",
    "title": "Sample English Clip",
    "duration": 8.5,
    "filename": "1768857488906-dga7itsuec-sample_clip.mp3",
    "originalFilename": "sample_clip.mp3",
    "fileSize": 204800,
    "metadata": {
      "language": "English",
      "speakerGender": "male",
      "speakerAgeRange": "adult",
      "speakerDialect": "American",
      "transcript": "This is a sample clip",
      "sourceUrl": "https://example.com/video",
      "tags": ["education", "podcast"]
    },
    "uploadedBy": "user-123",
    "createdAt": "2026-01-11T10:00:00Z",
    "updatedAt": "2026-01-11T10:00:00Z"
  }
]
```

### clips.csv Example

```csv
ID,Title,Duration(s),Language,Filename,UploadedBy,CreatedAt
"clip-1234567890-abc123","Sample English Clip",8.50,"English","1768857488906-dga7itsuec-sample_clip.mp3","user-123","2026-01-11T10:00:00Z"
```

## Examples

### Example 1: Get all clips metadata

```bash
./scraper.exe
```

Output:
```
🎵 Chorus Lab Scraper
====================

🔍 Fetching clips from https://the-chorusing-lab.vercel.app...
  📥 Fetching page 1 from: https://the-chorusing-lab.vercel.app/api/clips?limit=50&offset=0
  ✅ Retrieved 50 clips (total: 50)
  📥 Fetching page 2 from: https://the-chorusing-lab.vercel.app/api/clips?limit=50&offset=50
  ✅ Retrieved 30 clips (total: 80)
✅ Saved clips metadata: ./scraped_clips/clips.json
✅ Saved clips as CSV: ./scraped_clips/clips.csv

📊 Scraping Statistics:
  Total clips: 80
  Total duration: 456.30 minutes (7.61 hours)
  Languages:
    - English: 45 clips
    - Spanish: 25 clips
    - Urdu: 10 clips
  Speaker genders:
    - male: 42 clips
    - female: 35 clips
    - other: 3 clips

✨ Scraping complete!
📁 Output saved to: ./scraped_clips
```

### Example 2: Download English clips

```bash
./scraper.exe -language English -download -limit 20
```

### Example 3: Custom URL and directory

```bash
./scraper.exe -url http://localhost:3000 -output ./local_clips -download
```

## How It Works

1. **Fetch Clips**: Makes HTTP GET requests to `/api/clips` with pagination (50 clips per page)
2. **Parse JSON**: Unmarshals the API response into Go structs matching the TypeScript interfaces
3. **Download Audio**: For each clip, downloads the audio file from the API
4. **Save Metadata**: Exports clip information as JSON and CSV
5. **Statistics**: Calculates and displays summary statistics

## API Details

The scraper uses the `/api/clips` endpoint which expects:

**Query Parameters:**
- `offset` - Number of clips to skip
- `limit` - Number of clips to retrieve
- `language` - Filter by language
- Other filters: `speakerGender`, `speakerAgeRange`, `speakerDialect`, `uploadedBy`, `tags`, etc.

**Response Format:**
```json
{
  "clips": [
    {
      "id": "...",
      "title": "...",
      "duration": 8.5,
      "filename": "...",
      "metadata": { ... },
      "url": "/api/files/filename.mp3",
      ...
    }
  ]
}
```

## Troubleshooting

### Connection Error
```
❌ Error: failed to fetch clips: Get "https://...": dial tcp: lookup the-chorusing-lab.vercel.app: no such host
```
- Check your internet connection
- Verify the URL is correct: `-url https://the-chorusing-lab.vercel.app`

### Permission Denied
```
❌ Error: failed to create audio directory: permission denied
```
- Run with appropriate permissions or choose a different output directory
- On Windows: Right-click and "Run as Administrator"

### API Returns 404
```
❌ Error: API returned status 404
```
- The URL might be incorrect
- The API endpoint might not be available at that URL

### Out of Memory
- Use the `-limit` flag to scrape fewer clips at a time
- Run multiple scraping sessions with different limits

## Performance Tips

- Use `-limit` to test with a smaller number first
- The scraper uses 500ms delays between paginated requests to be respectful
- Audio files can be large; check available disk space before downloading
- Use `-language` filters to narrow down what you're scraping

## License

This scraper is provided as-is for the Chorus Lab project.

## Related Files

- `scraper.go` - The main Go script
- This README
- `src/types/audio.ts` - TypeScript interfaces that the scraper's structs match
- `src/app/api/clips/route.ts` - The API endpoint being scraped
