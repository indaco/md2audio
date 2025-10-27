# macOS say Provider

The `say` provider uses the built-in macOS text-to-speech system.

## Platform

- **macOS only**
- Built-in, no installation required

## Features

- Free (built-in with macOS)
- ~70 voices in various languages
- Multiple output formats (AIFF, M4A)
- Wide speaking rate range (90-360 WPM)
- Timing control support
- Offline operation

## Prerequisites

- macOS operating system
- No additional installation needed

## Setup

No setup required! The `say` command is built into macOS.

## Usage

### Basic Usage

```bash
# List available voices
./md2audio -provider say -list-voices

# Generate audio with default voice
./md2audio -f script.md

# Use voice preset
./md2audio -f script.md -p british-female

# Use specific voice
./md2audio -f script.md -v Kate
```

### Voice Presets

| Preset              | Voice Name |
|---------------------|------------|
| `british-female`    | Kate       |
| `british-male`      | Daniel     |
| `us-female`         | Samantha   |
| `us-male`           | Alex       |
| `australian-female` | Karen      |
| `indian-female`     | Veena      |

### Advanced Options

```bash
# Adjust speaking rate (lower = slower)
./md2audio -f script.md -v Kate -r 170

# Generate M4A instead of AIFF
./md2audio -f script.md -v Kate -format m4a

# Process entire directory
./md2audio -d ./docs -p british-female -o ./audio
```

## Output Formats

- **AIFF** (default) - Uncompressed, high quality
- **M4A** - Compressed, smaller file size

The tool uses AIFF internally and converts to M4A using `afconvert` if requested.

## Timing Control

The say provider supports timing annotations in H2 headers:

```markdown
## Introduction (8s)
This section will be adjusted to approximately 8 seconds.

## Main Content (5-10s)
This will target 10 seconds (uses the end time).
```

**How it works:**

- Counts words in the text
- Calculates required WPM to fit target duration
- Applies 0.95 adjustment factor for better accuracy
- Wide range: 90-360 WPM

**Accuracy:**

- Typical variance: 1-3 seconds from target
- Best for general-purpose narration
- Use `afinfo` to verify actual duration

## Common Voice Languages

Available voices include:

- **English**: US, UK, Australian, Indian, Irish, South African
- **Spanish**: Spain, Mexico, Argentina
- **French**: France, Canadian
- **German**: Germany
- **Italian**: Italy
- **Japanese**: Japan
- **Korean**: Korea
- **Chinese**: Mandarin, Cantonese
- And many more...

Run `./md2audio -list-voices` to see all available voices on your system.

## Tips

1. **Quality**: AIFF provides the best quality, M4A is more portable
2. **Rate**: Default is 180 WPM; adjust between 90-360 for different pacing
3. **Testing**: Use dry-run mode to preview before generating: `-dry-run`
4. **Caching**: Voice list is cached for 30 days for faster lookups

## Troubleshooting

### Voice not found

```bash
# List all available voices
./md2audio -provider say -list-voices

# Use exact voice name
./md2audio -f script.md -v "Samantha"
```

### Audio too fast/slow

```bash
# Slower speech (lower rate)
./md2audio -f script.md -v Kate -r 150

# Faster speech (higher rate)
./md2audio -f script.md -v Kate -r 200
```

### M4A conversion fails

- Ensure you have the latest macOS updates
- The `afconvert` command should be available by default
- Try generating AIFF first to verify the issue

## Performance

- Fast generation (local processing)
- No API rate limits
- Works offline
- Voice cache updates instantly

## Limitations

- macOS only (not available on Windows or Linux)
- Lower quality compared to Neural TTS services
- Limited voice customization (no pitch/volume control)
- Timing accuracy varies (±1-3 seconds typical)

## Next Steps

- Try [ElevenLabs](elevenlabs.md) for higher quality voices
- Try [Google Cloud TTS](google.md) for enterprise features
- Check [Provider Comparison](../provider-comparison.md) for detailed comparison
