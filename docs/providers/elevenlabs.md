# ElevenLabs Provider

The ElevenLabs provider uses the premium ElevenLabs AI text-to-speech API for highly realistic voice synthesis.

## Platform

- **Cross-platform** (Works on macOS, Linux, and Windows)
- Cloud-based API service

## Features

- Premium, highly realistic voices
- Emotional voice control
- Voice cloning capabilities
- Professional-grade quality
- Multilingual support
- Fine-grained voice settings
- Speed control (0.7x - 1.2x)
- Timing control support

## Prerequisites

- Any operating system (macOS, Linux, Windows)
- ElevenLabs API key ([Sign up here](https://elevenlabs.io/))
- Internet connection (cloud-based service)

## Setup

### 1. Get API Key

1. Sign up at [ElevenLabs](https://elevenlabs.io/)
2. Navigate to your profile settings
3. Copy your API key

### 2. Configure API Key

**Option A: Environment Variable**

```bash
export ELEVENLABS_API_KEY='your-api-key-here'
```

**Option B: .env File**

```bash
# Create .env file in your project directory
echo 'ELEVENLABS_API_KEY=your-api-key-here' > .env
```

**Option C: Command Line Flag**

```bash
./md2audio -provider elevenlabs \
  -elevenlabs-api-key your-api-key-here \
  -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM \
  -f script.md
```

### 3. (Optional) Configure Voice Settings

Fine-tune voice quality in `.env`:

```bash
ELEVENLABS_API_KEY=your-api-key-here
ELEVENLABS_STABILITY=0.5              # 0.0-1.0 (default: 0.5)
ELEVENLABS_SIMILARITY_BOOST=0.5       # 0.0-1.0 (default: 0.5)
ELEVENLABS_STYLE=0.0                  # 0.0-1.0 (default: 0.0)
ELEVENLABS_USE_SPEAKER_BOOST=true     # true/false (default: true)
ELEVENLABS_SPEED=1.0                  # 0.7-1.2 (default: 1.0)
```

## Usage

### List Available Voices

```bash
# List all ElevenLabs voices (cached)
./md2audio -provider elevenlabs -list-voices

# Refresh voice cache
./md2audio -provider elevenlabs -list-voices -refresh-cache

# Export voices to JSON
./md2audio -provider elevenlabs -export-voices voices.json
```

### Basic Generation

```bash
# Generate audio from markdown
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM \
  -f script.md

# Process entire directory
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id 21m00Tcm4TlvDq8ikWAM \
  -d ./docs \
  -o ./audio_output
```

### Using Specific Models

```bash
# Use multilingual model (default)
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id YOUR_VOICE_ID \
  -elevenlabs-model eleven_multilingual_v2 \
  -f script.md

# Use English-only model (lower latency)
./md2audio -provider elevenlabs \
  -elevenlabs-voice-id YOUR_VOICE_ID \
  -elevenlabs-model eleven_monolingual_v1 \
  -f script.md
```

## Voice Settings

### Stability (0.0-1.0)

Controls voice consistency:

- **Low (0.0-0.3)**: More expressive, variable
- **Medium (0.4-0.6)**: Balanced (default: 0.5)
- **High (0.7-1.0)**: Very consistent, less expressive

```bash
ELEVENLABS_STABILITY=0.7  # More consistent voice
```

### Similarity Boost (0.0-1.0)

Controls how closely the voice matches the original:

- **Low (0.0-0.3)**: More creative interpretation
- **Medium (0.4-0.6)**: Balanced (default: 0.5)
- **High (0.7-1.0)**: Closer to original voice characteristics

```bash
ELEVENLABS_SIMILARITY_BOOST=0.8  # Closer to original voice
```

### Style (0.0-1.0)

Controls emotional expression:

- **0.0**: No style/emotion (default, most stable)
- **0.1-0.5**: Subtle emotional variation
- **0.6-1.0**: High emotional expression

```bash
ELEVENLABS_STYLE=0.3  # Slight emotional variation
```

### Speaker Boost (true/false)

Enhances voice similarity:

- **true**: Better voice matching (default)
- **false**: Standard voice synthesis

```bash
ELEVENLABS_USE_SPEAKER_BOOST=true
```

### Speed (0.7-1.2)

Default speaking speed for non-timed sections:

- **0.7**: 30% slower
- **1.0**: Natural speed (default)
- **1.2**: 20% faster

```bash
ELEVENLABS_SPEED=1.1  # 10% faster for non-timed sections
```

**Note**: Sections with timing annotations (e.g., `## Intro (5s)`) automatically calculate speed to fit the duration.

## Timing Control

ElevenLabs supports timing annotations with automatic speed adjustment:

```markdown
## Introduction (8s)
This section will be adjusted to fit 8 seconds using speed control.

## Main Demo (5-10s)
Targets 10 seconds (end time is used).

## Conclusion
No timing specified - uses ELEVENLABS_SPEED setting (default: 1.0).
```

**Speed Range**: 0.7x - 1.2x

- **Accuracy**: Typically within 15% of target duration
- **Quality**: Natural-sounding speech maintained
- **Limitation**: If target requires speed outside range, it will be clamped with a warning

## Output Format

- **MP3 only** - ElevenLabs API returns MP3 audio
- High quality compression
- Suitable for all platforms

## Common Voice IDs

Popular ElevenLabs voices:

| Voice ID (2024)           | Name      | Description          |
|---------------------------|-----------|----------------------|
| 21m00Tcm4TlvDq8ikWAM      | Rachel    | Calm, professional   |
| AZnzlk1XvdvUeBnXmlld      | Domi      | Strong, confident    |
| EXAVITQu4vr4xnSDxMaL      | Bella     | Soft, friendly       |
| ErXwobaYiN019PkySvjV      | Antoni    | Well-rounded, male   |
| MF3mGyEYCl7XYWbV9V6O      | Elli      | Emotional, young     |
| TxGEqnHWrfWFTfGW9XjX      | Josh      | Deep, professional   |
| VR6AewLTigWG4xSOukaG      | Arnold    | Mature, authoritative|
| pNInz6obpgDQGcFmaJgB      | Adam      | Deep, narrative      |
| yoZ06aMxZJJ28mfd3POQ      | Sam       | Dynamic, energetic   |

Run `-list-voices` to see your account's available voices.

## Pricing

See [ElevenLabs Pricing](https://elevenlabs.io/pricing) for current details.

## Tips

1. **Start Simple**: Begin with default settings, then fine-tune
2. **Test First**: Generate one section to verify voice and settings
3. **Cache Voices**: First `-list-voices` call caches for 30 days
4. **Timing**: For tight timing needs, test and adjust markdown annotations
5. **Quality vs. Cost**: Higher quality settings may use more characters

## Troubleshooting

### API Key Errors

```bash
# Verify API key is set
echo $ELEVENLABS_API_KEY

# Or check .env file
cat .env | grep ELEVENLABS_API_KEY
```

### Voice Not Found

```bash
# List your available voices
./md2audio -provider elevenlabs -list-voices

# Copy the exact voice ID from the list
```

### Timing Issues

```bash
# Check calculated speed in output
# If speed is clamped (0.7 or 1.2), adjust target duration

# Example: If "Warning: Required speed (1.5) exceeds maximum"
# Increase target duration: (5s) → (7s)
```

## Performance

- **Latency**: Cloud-based, requires internet
- **Quality**: Premium, highly realistic
- **Rate Limits**: Depends on plan
- **Caching**: Voice list cached locally for 30 days
- **Retry Logic**: Automatic retry on transient failures

## Limitations

- MP3 format only (no WAV/AIFF)
- Requires internet connection
- API rate limits apply
- Speed range limited to 0.7x - 1.2x
- Costs scale with usage

## Best Practices

1. **Use .env**: Keep API keys out of scripts
2. **Cache Voices**: Run `-list-voices` once, then use cached list
3. **Batch Processing**: Process multiple files in one run
4. **Monitor Usage**: Check ElevenLabs dashboard regularly
5. **Test Settings**: Find optimal stability/similarity for your use case

## Next Steps

- Try [Google Cloud TTS](google.md) for even wider speed range
- Check [Provider Comparison](../provider-comparison.md) for detailed comparison
