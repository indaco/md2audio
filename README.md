<h1 align="center">
    md2audio
</h1>
<p align="center">
  <a href="https://github.com/indaco/md2audio/actions/workflows/ci.yml" target="_blank">
    <img src="https://github.com/indaco/md2audio/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://codecov.io/gh/indaco/md2audio">
    <img src="https://codecov.io/gh/indaco/md2audio/branch/main/graph/badge.svg" alt="Code coverage" />
  </a>
  <a href="https://goreportcard.com/report/github.com/indaco/md2audio/" target="_blank">
    <img src="https://goreportcard.com/badge/github.com/indaco/md2audio" alt="go report card" />
  </a>
  <a href="https://badge.fury.io/gh/indaco%2Fmd2audio">
    <img src="https://badge.fury.io/gh/indaco%2Fmd2audio.svg" alt="GitHub version" height="18">
  </a>
  <a href="https://pkg.go.dev/github.com/indaco/md2audio/" target="_blank">
      <img src="https://pkg.go.dev/badge/github.com/indaco/md2audio/.svg" alt="go reference" />
  </a>
   <a href="https://github.com/indaco/md2audio/blob/main/LICENSE" target="_blank">
    <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square&logo=none" alt="license" />
  </a>
  <a href="https://www.jetify.com/devbox/docs/contributor-quickstart/">
    <img src="https://www.jetify.com/img/devbox/shield_moon.svg" alt="Built with Devbox" />
  </a>
</p>

Convert markdown H2 sections to individual audio files using multiple TTS (Text-to-Speech) providers including macOS `say`, Linux `espeak-ng`, and ElevenLabs API.

## Features

- **Cross-Platform TTS Providers**: macOS `say`, Linux `espeak-ng`, and ElevenLabs API
- **Automatic Platform Detection**: Uses the best provider for your OS automatically
- **Process files or directories** recursively with structure mirroring
- **Target duration control**: Adjust timing with annotations like `(8s)`
- **Multiple formats**: AIFF, M4A, and MP3 output
- **Voice caching**: Fast lookups with SQLite WAL mode
- **Developer-friendly**: Debug mode, dry-run preview, progress indicators

## Prerequisites

### For macOS say Provider (Default on macOS)

- macOS (uses built-in `say` command)
- Go 1.25 or later (to build the tool)

### For Linux espeak Provider (Default on Linux)

- Linux (Ubuntu, Debian, Fedora, Arch, etc.)
- Go 1.25 or later (to build the tool)
- `espeak-ng` or `espeak` installed:

  ```bash
  # Ubuntu/Debian
  sudo apt install espeak-ng ffmpeg

  # Fedora/RHEL
  sudo dnf install espeak-ng ffmpeg

  # Arch Linux
  sudo pacman -S espeak-ng ffmpeg
  ```

- `ffmpeg` for audio format conversion (MP3, M4A support)

### For ElevenLabs Provider (Works on all platforms)

- Any OS (Windows, macOS, Linux)
- Go 1.25 or later (to build the tool)
- ElevenLabs API key ([Get one here](https://elevenlabs.io/))
- Set `ELEVENLABS_API_KEY` environment variable or create `.env` file

## Installation

### Using go install

```bash
go install github.com/indaco/md2audio/cmd/md2audio@latest
```

### Building from source

```bash
git clone https://github.com/indaco/md2audio.git
cd md2audio
go build -o md2audio ./cmd/md2audio
```

The binary will be created in the current directory. You can move it to a location in your PATH:

```bash
sudo mv md2audio /usr/local/bin/
```

## TTS Providers

md2audio supports multiple Text-to-Speech providers. The best provider for your platform is selected automatically:

### macOS say (Default on macOS)

- **Platform**: macOS only
- **Cost**: Free (built-in)
- **Setup**: No configuration needed
- **Quality**: Good for local development and testing
- **Formats**: AIFF, M4A
- **Voices**: ~70 voices in various languages

### Linux espeak-ng (Default on Linux)

- **Platform**: Linux only
- **Cost**: Free (open-source)
- **Setup**: Install `espeak-ng` and `ffmpeg`
- **Quality**: Good for local development and testing
- **Formats**: WAV, MP3, M4A, AIFF (via ffmpeg)
- **Voices**: 50+ voices in various languages
- **Voice Mapping**: Automatically maps macOS voice names (e.g., "Kate" → en-gb)

### ElevenLabs

- **Platform**: Cross-platform (works on any OS)
- **Cost**: Paid API ([Pricing](https://elevenlabs.io/pricing))
- **Setup**: Requires API key
- **Quality**: Premium, highly realistic voices
- **Formats**: MP3
- **Voices**: Multiple professional voices with emotional control

#### Setting up ElevenLabs

1. Get your API key from [ElevenLabs](https://elevenlabs.io/)

2. Set the environment variable:

   ```bash
   export ELEVENLABS_API_KEY='your-api-key'
   ```

3. Or create a `.env` file in your project directory:

   ```bash
   # Copy the example file
   cp .env.example .env
   # Then edit .env and add your API key
   ```

   Or create it directly:

   ```bash
   echo 'ELEVENLABS_API_KEY=your-api-key' > .env
   ```

4. (Optional) Configure voice settings in `.env`:

   ```bash
   # Voice quality settings (all optional, with sensible defaults)
   ELEVENLABS_STABILITY=0.5              # Voice consistency (0.0-1.0, default: 0.5)
   ELEVENLABS_SIMILARITY_BOOST=0.5       # Voice similarity (0.0-1.0, default: 0.5)
   ELEVENLABS_STYLE=0.0                  # Voice style/emotion (0.0-1.0, default: 0.0)
   ELEVENLABS_USE_SPEAKER_BOOST=true     # Boost similarity (true/false, default: true)
   ELEVENLABS_SPEED=1.0                  # Default speed for non-timed sections (0.7-1.2, default: 1.0)
   ```

   **Note:**
   - `ELEVENLABS_SPEED` only applies to sections WITHOUT timing annotations
   - Sections with `(5s)` timing will calculate speed automatically
   - Higher stability = more consistent but less expressive
   - Higher similarity_boost = closer to original voice characteristics
   - Style adds emotional range (0 = disabled, higher = more expressive)

5. List available voices:

   ```bash
   ./md2audio -provider elevenlabs -list-voices
   ```

## Usage

### Basic Examples

#### Using Default Provider (say on macOS, espeak on Linux)

```bash
# Check version
./md2audio -version

# List available voices (automatically uses the best provider for your OS)
./md2audio -list-voices

# Process a single markdown file with voice preset
# Works on both macOS (say) and Linux (espeak) automatically!
./md2audio -f script.md -p british-female

# Process entire directory recursively
./md2audio -d ./docs -p british-female

# Use specific voice with slower rate for clarity
# On macOS: uses "Kate" voice directly
# On Linux: maps "Kate" to "en-gb" voice automatically
./md2audio -f script.md -v Kate -r 170

# Generate M4A files instead of default format
# macOS default: AIFF, Linux default: WAV
./md2audio -d ./content -p british-female -format m4a

# Custom output directory and prefix
./md2audio -f script.md -o ./voiceovers -prefix demo

# Preview what would be generated (dry-run mode)
./md2audio -f script.md -p british-female -dry-run

# Enable debug logging to troubleshoot issues
./md2audio -f script.md -p british-female -debug

# Combine dry-run with debug for detailed preview
./md2audio -d ./docs -p british-female -dry-run -debug

# Explicitly use espeak provider (on any Linux system)
./md2audio -f script.md -provider espeak -v en-gb

# Explicitly use say provider (on macOS)
./md2audio -f script.md -provider say -v Kate
```

#### Using ElevenLabs Provider

```bash
# List available ElevenLabs voices (cached for faster access)
./md2audio -provider elevenlabs -list-voices

# Refresh voice cache (when new voices are available)
./md2audio -provider elevenlabs -list-voices -refresh-cache

# Export voices to JSON for reference
./md2audio -provider elevenlabs -export-voices elevenlabs_voices.json

# Process a single file with ElevenLabs
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM \
  -f script.md

# Process entire directory with ElevenLabs
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM \
  -d ./docs \
  -o ./audio_output

# Use specific ElevenLabs model
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id YOUR_VOICE_ID \
  -elevenlabs-model eleven_multilingual_v2 \
  -f script.md
```

### Debug Mode

Enable debug logging to troubleshoot issues or understand what's happening under the hood:

```bash
# Enable debug logging
./md2audio -f script.md -p british-female -debug
```

**Debug mode shows:**

- Cache hits/misses for voice lookups
- API request details (ElevenLabs)
- File processing progress
- Internal operation details

**When to use debug mode:**

- Troubleshooting API issues with ElevenLabs
- Understanding cache behavior
- Investigating performance problems
- Reporting bugs with detailed logs

### Dry-Run Mode

Preview what would be generated without creating any audio files:

```bash
# Dry-run mode - shows what would be generated
./md2audio -f script.md -p british-female -dry-run

# Combine with debug for maximum visibility
./md2audio -d ./docs -provider elevenlabs -elevenlabs-voice-id YOUR_ID -dry-run -debug
```

**Dry-run mode shows:**

- Which sections would be processed
- Output file paths that would be created
- Timing information for timed sections
- Preview of text content

**When to use dry-run mode:**

- Testing markdown format before generation
- Verifying output paths and filenames
- Checking section count and structure
- Planning batch processing jobs

**Example output:**

```
💡 DRY-RUN MODE: No files will be created

ℹ Section 1/3:
  - title: Introduction
  💡 Target duration: 8.0 seconds
  💡 Text: Welcome to this demonstration...
  Would create: ./audio_sections/section_01_introduction.aiff

ℹ Section 2/3:
  - title: Main Content
  💡 Text: Here is the main content...
  Would create: ./audio_sections/section_02_main_content.aiff

✔ Would generate 3 audio files
```

### Voice Caching

To improve performance, md2audio caches voice lists from providers. This is especially useful for ElevenLabs to avoid repeated API calls:

```bash
# First call - fetches from API and caches (slower)
./md2audio -provider elevenlabs -list-voices

# Subsequent calls - uses cache (instant)
./md2audio -provider elevenlabs -list-voices

# Force refresh when new voices are available
./md2audio -provider elevenlabs -list-voices -refresh-cache

# Export cached voices to JSON file for reference
./md2audio -provider elevenlabs -export-voices elevenlabs_voices.json
./md2audio -provider say -export-voices say_voices.json
```

**Cache Details:**

- **Location**: `~/.md2audio/voice_cache.db` (SQLite database)
- **Duration**: 30 days (voices don't change frequently)
- **Benefits**: Instant voice listing, reduced API calls, offline access to voice list
- **Refresh**: Use `-refresh-cache` flag when you know new voices are available

### Command Line Options

#### General Options

| Flag             | Description                                         | Default                 |
| ---------------- | --------------------------------------------------- | ----------------------- |
| `-f`             | Input markdown file (use `-f` or `-d`)              | -                       |
| `-d`             | Input directory (recursive, use `-f` or `-d`)       | -                       |
| `-o`             | Output directory                                    | `./audio_sections`      |
| `-format`        | Output format                                       | `aiff`                  |
| `-prefix`        | Filename prefix                                     | `section`               |
| `-list-voices`   | List all available voices (uses cache if available) | -                       |
| `-refresh-cache` | Force refresh of voice cache                        | `false`                 |
| `-export-voices` | Export cached voices to JSON file                   | -                       |
| `-provider`      | TTS provider (`say`, `espeak`, or `elevenlabs`)     | Auto-detect by platform |
| `-version`       | Print version and exit                              | -                       |
| `-debug`         | Enable debug logging                                | `false`                 |
| `-dry-run`       | Show what would be generated without creating files | `false`                 |

#### say/espeak Provider Options

These options work for both `say` (macOS) and `espeak` (Linux) providers:

| Flag | Description                            | Default             |
| ---- | -------------------------------------- | ------------------- |
| `-p` | Voice preset (see Voice Presets below) | `Kate` (if not set) |
| `-v` | Specific voice name (overrides `-p`)   | -                   |
| `-r` | Speaking rate (lower = slower)         | `180`               |

**Note:** Voice names are automatically mapped between platforms. For example, "Kate" uses the Kate voice on macOS and en-gb on Linux.

#### ElevenLabs Provider Options

| Flag                   | Description                         | Default                  |
| ---------------------- | ----------------------------------- | ------------------------ |
| `-elevenlabs-voice-id` | ElevenLabs voice ID (required)      | -                        |
| `-elevenlabs-model`    | ElevenLabs model ID                 | `eleven_multilingual_v2` |
| `-elevenlabs-api-key`  | ElevenLabs API key (prefer env var) | `ELEVENLABS_API_KEY` env |

### Voice Presets

These presets work on both macOS and Linux (automatically mapped):

| Preset              | macOS Voice | Linux Voice |
| ------------------- | ----------- | ----------- |
| `british-female`    | Kate        | en-gb       |
| `british-male`      | Daniel      | en-gb       |
| `us-female`         | Samantha    | en-us       |
| `us-male`           | Alex        | en-us       |
| `australian-female` | Karen       | en-au       |
| `indian-female`     | Veena       | en-in       |

**Cross-Platform Usage:**

```bash
# Same command works on both macOS and Linux!
./md2audio -f script.md -p british-female

# Or use specific voices (automatically mapped)
./md2audio -f script.md -v Kate  # macOS: Kate, Linux: en-gb
```

### ElevenLabs Voice Settings

ElevenLabs voice quality can be fine-tuned using environment variables. All settings are optional and have sensible defaults:

| Setting                        | Range      | Default | Description                                                            |
| ------------------------------ | ---------- | ------- | ---------------------------------------------------------------------- |
| `ELEVENLABS_STABILITY`         | 0.0-1.0    | 0.5     | Voice consistency. Higher = more consistent but less expressive        |
| `ELEVENLABS_SIMILARITY_BOOST`  | 0.0-1.0    | 0.5     | Voice similarity to original. Higher = closer to voice characteristics |
| `ELEVENLABS_STYLE`             | 0.0-1.0    | 0.0     | Emotional range. 0 = disabled, higher = more expressive                |
| `ELEVENLABS_USE_SPEAKER_BOOST` | true/false | true    | Boost similarity of synthesized speech                                 |
| `ELEVENLABS_SPEED`             | 0.7-1.2    | 1.0     | Default speaking speed (only for sections without timing annotations)  |

**Speed Behavior:**

- Sections **with** timing annotations like `## Scene 1 (5s)` → Speed is calculated automatically to fit duration
- Sections **without** timing annotations → Uses `ELEVENLABS_SPEED` setting (default: 1.0)

**Example `.env` configuration:**

```bash
ELEVENLABS_API_KEY=your-api-key
ELEVENLABS_STABILITY=0.7           # More consistent voice
ELEVENLABS_SIMILARITY_BOOST=0.8    # Closer to original voice
ELEVENLABS_STYLE=0.3               # Slight emotional variation
ELEVENLABS_SPEED=1.1               # 10% faster for non-timed sections
```

## Markdown Format

The script expects H2 headers (`##`) to denote sections. You can optionally specify target duration for each section:

```markdown
## Scene 1: Introduction (8s)

This is the content for scene 1. It will be converted to audio that lasts exactly 8 seconds.

## Scene 2: Main Demo (12s)

This is the content for scene 2. The speaking rate will be automatically adjusted to fit 12 seconds.

## Scene 3: Conclusion

This section has no timing specified, so it will use the default speaking rate (-r flag).
```

### Timing Formats Supported

- `(8s)` - Target duration of 8 seconds
- `(10.5s)` - Target duration of 10.5 seconds
- `(0-8s)` - Range format, uses end time (8 seconds)
- `(15 seconds)` - Also works with "seconds" spelled out

**How it works (macOS say provider only):**

- The script counts the words in your text
- Calculates the required words-per-minute (WPM) to fit the target duration
- Automatically adjusts the speaking rate for that section
- Shows you the actual duration vs target after generation

**Important Notes:**

- **Timing is supported with both providers**, but with different accuracy:
  - **macOS say provider**: Uses `-r` (rate) parameter for speed control
    - Very wide range of speaking rates (90-360 wpm)
    - Actual duration may differ from target (typically within 1-3 seconds)
    - Applies 0.95 adjustment factor for better accuracy

  - **ElevenLabs provider**: Uses `speed` parameter (NEW!)
    - Limited range: 0.7x (slower) to 1.2x (faster) of natural pace
    - More accurate natural-sounding speech
    - If target duration requires speed outside this range, audio will be clamped
    - Example: 5s target → 5.75s actual (within 15% for typical content)

- **Timing accuracy tip**: Test with your content and adjust timing annotations as needed. For very tight timing requirements, consider the say provider's wider speed range.

## Directory Processing

Process entire directory trees recursively with the `-d` flag:

```bash
./md2audio -d ./docs -p british-female -o ./audio_output
```

**Input structure:**

```
docs/
├── intro.md
├── chapter1/
│   ├── part1.md
│   └── part2.md
└── chapter2/
    └── overview.md
```

**Output structure (mirrors input):**

```
audio_output/
├── intro/
│   ├── section_01_welcome.aiff
│   └── section_02_overview.aiff
├── chapter1/
│   ├── part1/
│   │   ├── section_01_title.aiff
│   │   └── section_02_title.aiff
│   └── part2/
│       └── section_01_title.aiff
└── chapter2/
    └── overview/
        └── section_01_title.aiff
```

**Key features:**

- Processes all `.md` files recursively
- Creates mirror directory structure
- Each markdown file gets its own subdirectory
- Preserves folder hierarchy from input
- Continues processing even if individual files fail

**Example with examples folder:**

```bash
# Process the included examples
./md2audio -d ./examples -p british-female -format m4a

# Results in organized audio files matching the examples structure
```

## Output

Files are named using the pattern:

```
{prefix}_{number}_{sanitized_title}.{format}
```

Example outputs:

- `section_01_scene_1_introduction.aiff`
- `section_02_scene_2_main_demo.aiff`

## Tips for Video Editing

1. Generate separate files per section (this is automatic)
2. Add timing to your markdown headers to match your screen recording
3. Import all audio files into your video editing software
4. Place each audio clip on the timeline where needed
5. The audio will match your specified durations automatically

### Timing Tips

- **Be realistic**: Very short durations with lots of text will sound rushed
- **Test first**: Generate one section to verify the pacing feels natural
- **Adjust if needed**: If timing is off, adjust the duration in your markdown and regenerate
- **Word count matters**: ~2-3 words per second is natural speech
- **Override if needed**: The `-r` flag still works for sections without timing

## Troubleshooting

**Voice not found:**

- Run `./md2audio -list-voices` to see available voices
- Use the exact voice name with `-v` flag

**No sections found:**

- Ensure your markdown uses `##` for headers (H2)
- Check there's content after each header

**Audio quality:**

- AIFF format is higher quality but larger
- M4A format is compressed and smaller
- Adjust rate with `-r` flag for clarity

## Example Workflow

```bash
# 1. Check your markdown format
cat examples/demo_script.md

# 2. List available voices
./md2audio -list-voices

# 3. Generate audio files
./md2audio -f examples/demo_script.md -p british-female -r 175 -format m4a

# 4. Import the files from ./audio_sections into your video editor
```

## Notes

- The script automatically cleans markdown formatting (links, bold, italic)
- Empty sections are skipped
- Section titles are sanitized for safe filenames
- Speaking rate default is 180 (macOS default is 200)

## For Developers

Interested in contributing or understanding the codebase?

See the [Contributing Guide](CONTRIBUTING.md) for detailed information about:

- Project architecture and package organization
- Development tools and workflow
- Code quality standards
- Setting up your development environment

## Contributing

Contributions are welcome!

See the [Contributing Guide](/CONTRIBUTING.md) for setup instructions.

## License

This project is licensed under the MIT License – see the [LICENSE](./LICENSE) file for details.
