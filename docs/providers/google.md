# Google Cloud TTS Example

This example demonstrates using Google Cloud Text-to-Speech with md2audio.

## Setup

### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Cloud Text-to-Speech API
4. Create a service account:
   - Go to IAM & Admin > Service Accounts
   - Click "Create Service Account"
   - Grant "Cloud Text-to-Speech User" role
   - Create and download a JSON key file

### 2. Configure Credentials

Set the environment variable to point to your service account key:

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/service-account-key.json"
```

Or add it to your `.env` file:

```bash
echo 'GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json' > .env
```

## Usage Examples

### List Available Voices

```bash
# List all Google Cloud TTS voices (400+ voices)
./md2audio -provider google -list-voices

# Export voices to JSON file for reference
./md2audio -provider google -export-voices google-voices.json
```

### Generate Audio from Markdown

```bash
# Process a single file with Neural2 voice (high quality)
./md2audio -provider google -google-voice en-US-Neural2-F -f examples/demo_script.md

# Process with British English voice
./md2audio -provider google -google-voice en-GB-Neural2-A -f examples/demo_script.md

# Generate MP3 files
./md2audio -provider google -google-voice en-US-Neural2-F -format mp3 -f examples/demo_script.md

# Process entire directory
./md2audio -provider google -google-voice en-US-Neural2-C -d ./docs -o ./audio_output
```

### Advanced Options

```bash
# Adjust speaking rate (0.25 = very slow, 4.0 = very fast)
./md2audio -provider google -google-voice en-US-Neural2-F -google-speed 1.5 -f script.md

# Adjust pitch (-20.0 to 20.0 semitones)
./md2audio -provider google -google-voice en-US-Neural2-F -google-pitch 2.0 -f script.md

# Adjust volume gain (-96.0 to 16.0 dB)
./md2audio -provider google -google-voice en-US-Neural2-F -google-volume 3.0 -f script.md

# Use different language
./md2audio -provider google -google-voice es-ES-Neural2-A -google-language es-ES -f spanish_script.md
```

## Voice Types

Google Cloud TTS offers multiple voice quality tiers:

### Neural2 (Recommended - Best Quality)

- `en-US-Neural2-F` - Female, American English
- `en-US-Neural2-M` - Male, American English
- `en-GB-Neural2-A` - Female, British English
- `en-GB-Neural2-B` - Male, British English

### WaveNet (High Quality)

- `en-US-Wavenet-F` - Female, American English
- `en-US-Wavenet-A` - Male, American English

### Standard (Basic Quality)

- `en-US-Standard-A` - Female, American English
- `en-US-Standard-D` - Male, American English

### Studio (Premium Quality - Highest Fidelity)

- `en-US-Studio-M` - Male voice optimized for studio recordings
- `en-US-Studio-O` - Female voice optimized for studio recordings

## Output Formats

Google Cloud TTS supports:

- **MP3** - Compressed, good for web use
- **WAV** - Uncompressed, high quality (LINEAR16 encoding)
- **OGG** - Compressed with Opus codec

```bash
# Generate WAV files
./md2audio -provider google -google-voice en-US-Neural2-F -format wav -f script.md

# Generate OGG files
./md2audio -provider google -google-voice en-US-Neural2-F -format ogg -f script.md
```

## Timing Control

Google Cloud TTS has the widest speaking rate range (0.25x - 4.0x):

```bash
# Slow speech for learning materials
./md2audio -provider google -google-voice en-US-Neural2-F -google-speed 0.75 -f lesson.md

# Fast speech for quick reviews
./md2audio -provider google -google-voice en-US-Neural2-F -google-speed 1.5 -f summary.md
```

The tool also supports timing annotations in H2 headers:

```markdown
## Introduction (8s)
This section will be adjusted to speak in approximately 8 seconds.

## Quick Overview (3-5s)
This will be between 3 and 5 seconds long.
```

## Multi-Language Support

Google Cloud TTS supports 50+ languages:

```bash
# Spanish
./md2audio -provider google -google-voice es-ES-Neural2-A -google-language es-ES -f spanish.md

# French
./md2audio -provider google -google-voice fr-FR-Neural2-A -google-language fr-FR -f french.md

# German
./md2audio -provider google -google-voice de-DE-Neural2-F -google-language de-DE -f german.md

# Japanese
./md2audio -provider google -google-voice ja-JP-Neural2-B -google-language ja-JP -f japanese.md
```

## Pricing

See [Google Cloud TTS Pricing](https://cloud.google.com/text-to-speech/pricing) for current rates.

## Tips

1. **Voice Selection**: Start with Neural2 voices for the best quality-to-cost ratio
2. **Caching**: The first `-list-voices` call downloads all voices; subsequent calls use cache (instant)
3. **Credentials**: Keep your service account key secure, never commit it to version control
4. **IAM Permissions**: Ensure your service account has the "Cloud Text-to-Speech User" role
5. **Rate Limits**: Google Cloud TTS has generous rate limits, but for bulk processing consider batching
6. **Language Codes**: Use the same language code in voice name and `-google-language` flag

## Troubleshooting

### "Google Cloud credentials not found" error

- Ensure `GOOGLE_APPLICATION_CREDENTIALS` is set correctly
- Check that the service account key file exists and is readable
- Verify the path doesn't contain typos

### "Permission denied" error

- Check that your service account has the "Cloud Text-to-Speech User" role
- Ensure the Cloud Text-to-Speech API is enabled in your project

### Voices not appearing

- Run with `-refresh-cache` to force update the voice cache
- Check your internet connection
- Verify API access from your network

## Resources

- [Google Cloud TTS Documentation](https://cloud.google.com/text-to-speech/docs)
- [Voice List](https://cloud.google.com/text-to-speech/docs/voices)
- [SSML Support](https://cloud.google.com/text-to-speech/docs/ssml) (future feature)
- [Audio Profiles](https://cloud.google.com/text-to-speech/docs/audio-profiles) (future feature)
- Check [Provider Comparison](../provider-comparison.md) for detailed comparison
