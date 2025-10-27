# md2audio

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

<h2 align="center">Convert Markdown H2 sections to individual audio files using multiple TTS providers.</h2>

> [!WARNING]
> This project is under active development. You may encounter bugs or incomplete features. Please report any issues on the [GitHub issue tracker](https://github.com/indaco/md2audio/issues).

## Features

- **Multiple TTS Providers**: Choose from macOS say, Linux espeak, Google Cloud TTS, or ElevenLabs
- **Cross-Platform**: Works on macOS, Linux, and Windows (with cloud providers)
- **Automatic Platform Detection**: Uses the best provider for your OS by default
- **Timing Control**: Specify target durations with annotations like `(8s)`
- **Batch Processing**: Process files or entire directories recursively
- **Voice Caching**: Fast voice lookups with SQLite-based caching
- **Multiple Formats**: AIFF, M4A, MP3, WAV, OGG output formats
- **Developer Tools**: Debug mode, dry-run preview, progress indicators

## Installation

### 1. Global Install (via go install)

```bash
go install github.com/indaco/md2audio/cmd/md2audio@latest
```

### 2. Prebuilt binaries

Download the pre-compiled binaries from the [releases page](https://github.com/md2audio/tempo/releases) and move the binary to a folder in your system's PATH.

### 3. Build from Source

```bash
git clone https://github.com/indaco/md2audio.git
cd md2audio
go build -o md2audio ./cmd/md2audio # move the binary to a folder in your system's PATH
```

or with [just](https://just.systems/man/en/)

```bash
just install
```

## Basic Usage

```bash
# Process a markdown file (uses default provider for your OS)
./md2audio -f script.md -p british-female

# Process entire directory
./md2audio -d ./docs -p british-female -o ./audio

# List available voices
./md2audio -list-voices
```

## TTS Providers

md2audio supports multiple text-to-speech providers. Choose the one that best fits your needs:

| Provider                                         | Platform | Cost | Quality | Best For                       |
| ------------------------------------------------ | -------- | ---- | ------- | ------------------------------ |
| **[say](docs/providers/say.md)**                 | macOS    | Free | Good    | Local dev/testing              |
| **[espeak](docs/providers/espeak.md)**           | Linux    | Free | Basic   | Linux dev/testing              |
| **[Google Cloud TTS](docs/providers/google.md)** | All      | Paid | Premium | Enterprise, multi-language     |
| **[ElevenLabs](docs/providers/elevenlabs.md)**   | All      | Paid | Premium | Production content, audiobooks |

**[Compare Providers](docs/provider-comparison.md)** - Detailed comparison to help you choose

### Quick Provider Examples

```bash
# macOS say (default on macOS)
./md2audio -f script.md -p british-female

# Linux espeak (default on Linux)
./md2audio -f script.md -provider espeak -v en-gb

# Google Cloud TTS
./md2audio -provider google -google-voice en-US-Neural2-F -f script.md

# ElevenLabs
./md2audio -provider elevenlabs -elevenlabs-voice-id VOICE_ID -f script.md
```

## Markdown Format

Use H2 headers (`##`) to denote sections. Add optional timing annotations:

```markdown
## Introduction (8s)

This section will be adjusted to approximately 8 seconds.

## Main Content (5-10s)

This targets 10 seconds (end time is used).

## Conclusion

No timing specified - uses default speaking rate.
```

**Supported timing formats**: `(8s)`, `(10.5s)`, `(0-8s)`, `(15 seconds)`

## Command Line Options

### General Options

| Flag           | Description                                            | Default                        |
| -------------- | ------------------------------------------------------ | ------------------------------ |
| `-f`           | Input markdown file                                    | -                              |
| `-d`           | Input directory (recursive)                            | -                              |
| `-o`           | Output directory                                       | `./audio_sections`             |
| `-provider`    | TTS provider (`say`, `espeak`, `elevenlabs`, `google`) | Auto-detect                    |
| `-format`      | Output format (`aiff`, `m4a`, `mp3`, `wav`, `ogg`)     | `aiff` (macOS) / `wav` (Linux) |
| `-prefix`      | Filename prefix                                        | `section`                      |
| `-list-voices` | List available voices                                  | -                              |
| `-version`     | Print version                                          | -                              |
| `-debug`       | Enable debug logging                                   | `false`                        |
| `-dry-run`     | Preview without generating files                       | `false`                        |

### Provider-Specific Options

Each provider has its own configuration options. See the provider guides for details:

- **say/espeak**: `-p` (voice preset), `-v` (voice name), `-r` (speaking rate)
- **Google Cloud**: `-google-voice`, `-google-language`, `-google-credentials`, `-google-speed`, `-google-pitch`, `-google-volume`
- **ElevenLabs**: `-elevenlabs-voice-id`, `-elevenlabs-model`, `-elevenlabs-api-key` (voice settings via env vars)

**[Provider Documentation](docs/providers/)** for complete option lists

## Examples

### Basic Examples

```bash
# Preview what would be generated
./md2audio -f script.md -p british-female -dry-run

# Generate M4A files instead of AIFF
./md2audio -f script.md -p british-female -format m4a

# Process directory with custom output location
./md2audio -d ./content -p us-female -o ./voiceovers

# Enable debug logging
./md2audio -f script.md -debug
```

### Provider-Specific Examples

```bash
# Google Cloud TTS with Neural2 voice
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/creds.json"
./md2audio -provider google \
  -google-voice en-US-Neural2-F \
  -format mp3 \
  -d ./docs

# ElevenLabs with custom settings
export ELEVENLABS_API_KEY='your-key'
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM \
  -elevenlabs-model eleven_multilingual_v2 \
  -f script.md

# espeak on Linux with MP3 output
./md2audio -provider espeak \
  -v en-gb \
  -format mp3 \
  -d ./docs
```

## Voice Caching

md2audio caches voice lists locally for faster access:

```bash
# First call - fetches from provider and caches
./md2audio -provider google -list-voices

# Subsequent calls - uses cache (instant)
./md2audio -provider google -list-voices

# Force refresh when new voices available
./md2audio -provider google -list-voices -refresh-cache

# Export to JSON for reference
./md2audio -provider google -export-voices voices.json
```

- **Cache Location**: `~/.md2audio/voice_cache.db`
- **Cache Duration**: 30 days
- **Supported Providers**: All providers

## Output Structure

### Single File

```bash
./md2audio -f script.md -p british-female
```

Output:

```
audio_sections/
├── section_01_introduction.aiff
├── section_02_main_content.aiff
└── section_03_conclusion.aiff
```

### Directory Processing

```bash
./md2audio -d ./docs -p british-female
```

Input structure is mirrored in output:

```
docs/                          audio_sections/
├── intro.md              →    ├── intro/
├── chapter1/             →    ├── chapter1/
│   ├── part1.md          →    │   ├── part1/
│   └── part2.md          →    │   └── part2/
└── chapter2/             →    └── chapter2/
    └── overview.md       →        └── overview/
```

## Troubleshooting

### Voice Not Found

```bash
# List all available voices for your provider
./md2audio -list-voices

# Use exact voice name from the list
./md2audio -f script.md -v "Samantha"
```

### Provider Setup Issues

See the provider-specific guide for detailed setup instructions:

- [say Setup](docs/providers/say.md#setup) - No setup needed
- [espeak Setup](docs/providers/espeak.md#installation) - Install espeak-ng
- [Google Cloud Setup](docs/providers/google.md#setup) - GCP credentials
- [ElevenLabs Setup](docs/providers/elevenlabs.md#setup) - API key

### Debug Mode

Enable debug logging to troubleshoot issues:

```bash
./md2audio -f script.md -p british-female -debug
```

Shows:

- Cache hits/misses
- API request details
- File processing progress
- Internal operations

## Contributing

Contributions are welcome! See the [Contributing Guide](CONTRIBUTING.md) for:

- Project architecture
- Development setup
- Code quality standards
- Provider implementation guide

## License

MIT License - see [LICENSE](LICENSE) for details
