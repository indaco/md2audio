# Linux espeak Provider

The `espeak` provider uses the open-source eSpeak NG text-to-speech synthesizer for Linux systems.

## Platform

- **Linux only** (Ubuntu, Debian, Fedora, Arch, etc.)
- Open source and free

## Features

- Free and open source
- 50+ voices in various languages
- Multiple output formats (WAV, MP3, M4A, AIFF via ffmpeg)
- Timing control support
- Offline operation
- Cross-platform voice mapping (macOS voice names work automatically)

## Prerequisites

- Linux operating system
- `espeak-ng` or `espeak` installed
- `ffmpeg` for format conversion (MP3, M4A, AIFF)

## Installation

### Ubuntu/Debian

```bash
sudo apt install espeak-ng ffmpeg
```

### Fedora/RHEL

```bash
sudo dnf install espeak-ng ffmpeg
```

### Arch Linux

```bash
sudo pacman -S espeak-ng ffmpeg
```

## Setup

After installation, verify espeak is available:

```bash
# Check espeak-ng is installed
which espeak-ng

# Or check for espeak (older version)
which espeak

# Test voice
espeak-ng "Hello, this is a test"
```

## Usage

### Basic Usage

```bash
# List available voices
./md2audio -provider espeak -list-voices

# Generate audio with default voice
./md2audio -f script.md

# Use voice preset (automatically mapped)
./md2audio -f script.md -p british-female

# Use specific voice
./md2audio -f script.md -v en-gb
```

### Voice Presets (Cross-Platform)

These presets work the same on Linux and macOS:

| Preset              | macOS Voice | Linux Voice (espeak) |
|---------------------|-------------|----------------------|
| `british-female`    | Kate        | en-gb                |
| `british-male`      | Daniel      | en-gb                |
| `us-female`         | Samantha    | en-us                |
| `us-male`           | Alex        | en-us                |
| `australian-female` | Karen       | en-au                |
| `indian-female`     | Veena       | en-in                |

**Cross-platform example:**

```bash
# Same command works on both macOS and Linux!
./md2audio -f script.md -p british-female

# macOS voice names are automatically mapped
./md2audio -f script.md -v Kate  # Becomes en-gb on Linux
```

### Advanced Options

```bash
# Adjust speaking rate (lower = slower)
./md2audio -f script.md -v en-gb -r 170

# Generate MP3 instead of WAV
./md2audio -f script.md -v en-gb -format mp3

# Generate M4A
./md2audio -f script.md -v en-gb -format m4a

# Process entire directory
./md2audio -d ./docs -p british-female -o ./audio
```

## Output Formats

- **WAV** (default) - Uncompressed, high quality
- **MP3** - Compressed, good quality (requires ffmpeg)
- **M4A** - Compressed, compatible with Apple devices (requires ffmpeg)
- **AIFF** - Uncompressed, Apple format (requires ffmpeg)

## Timing Control

The espeak provider supports timing annotations in H2 headers:

```markdown
## Introduction (8s)
This section will be adjusted to approximately 8 seconds.

## Main Content (5-10s)
This will target 10 seconds (uses the end time).
```

**How it works:**

- Similar to macOS say provider
- Adjusts speaking rate to fit target duration
- Uses words-per-minute calculation

## Common Voice Languages

Available voices include:

- **English**: US (en-us), UK (en-gb), Australian (en-au), etc.
- **Spanish**: es, es-la (Latin America)
- **French**: fr, fr-be (Belgian)
- **German**: de
- **Italian**: it
- **Portuguese**: pt, pt-br (Brazilian)
- **Russian**: ru
- **Chinese**: zh (Mandarin)
- **Japanese**: ja
- And many more...

Run `./md2audio -provider espeak -list-voices` to see all available voices.

## Voice Mapping

When you use macOS voice names on Linux, they're automatically mapped:

| macOS Voice | Linux espeak Voice |
|-------------|-------------------|
| Kate        | en-gb             |
| Daniel      | en-gb             |
| Samantha    | en-us             |
| Alex        | en-us             |
| Karen       | en-au             |
| Veena       | en-in             |

This allows scripts and commands to work across both platforms without modification.

## Tips

1. **Quality**: WAV provides lossless quality, MP3 is more portable
2. **ffmpeg**: Required for MP3, M4A, and AIFF output formats
3. **Testing**: Use dry-run mode to preview: `-dry-run`
4. **Caching**: Voice list is cached for 30 days for faster lookups
5. **Cross-platform**: Use voice presets for portable scripts

## Troubleshooting

### espeak-ng not found

```bash
# Ubuntu/Debian
sudo apt install espeak-ng

# Fedora
sudo dnf install espeak-ng

# Arch
sudo pacman -S espeak-ng
```

### Format conversion fails

```bash
# Install ffmpeg for MP3/M4A/AIFF support
sudo apt install ffmpeg  # Ubuntu/Debian
sudo dnf install ffmpeg  # Fedora
sudo pacman -S ffmpeg    # Arch
```

### Voice not found

```bash
# List all available voices
./md2audio -provider espeak -list-voices

# Use espeak voice code
./md2audio -f script.md -v en-gb
```

### Audio quality issues

- espeak-ng generally has better quality than legacy espeak
- For higher quality, consider ElevenLabs or Google Cloud TTS
- Adjust rate for clearer speech: `-r 170`

## Performance

- Fast generation (local processing)
- No API rate limits
- Works offline
- Voice cache updates instantly
- Lightweight resource usage

## Limitations

- Linux only (not available on macOS or Windows)
- Robotic voice quality (not neural TTS)
- Limited voice customization
- Timing accuracy varies

## Next Steps

- [ElevenLabs](elevenlabs.md) - Cloud-based, premium quality
- [Google Cloud TTS](google.md) - Enterprise features, Neural2 voices
- Check [Provider Comparison](../provider-comparison.md) for detailed comparison
