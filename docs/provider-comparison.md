# TTS Provider Comparison

This guide helps you choose the best text-to-speech provider for your needs.

## Quick Comparison Table

| Feature         | [say](providers/say.md) | [espeak](providers/espeak.md) | [ElevenLabs](providers/elevenlabs.md) | [Google Cloud](providers/google.md) |
| --------------- | ----------------------- | ----------------------------- | ------------------------------------- | ----------------------------------- |
| **Platform**    | macOS only              | Linux only                    | All platforms                         | All platforms                       |
| **Cost**        | Free                    | Free                          | Paid ($5-99/mo)                       | Paid ($4-16/M chars)                |
| **Quality**     | Good                    | Basic                         | Premium                               | Premium                             |
| **Voices**      | ~70 voices              | ~50 voices                    | 20+ premium                           | 400+ voices                         |
| **Languages**   | 30+                     | 50+                           | 30+                                   | 50+                                 |
| **Offline**     | Yes                     | Yes                           | ❌ No                                 | ❌ No                               |
| **Speed Range** | 90-360 WPM              | Variable                      | 0.7x-1.2x                             | 0.25x-4.0x                          |
| **Formats**     | AIFF, M4A               | WAV, MP3, M4A                 | MP3                                   | MP3, WAV, OGG                       |
| **Setup**       | None                    | Install espeak                | API key                               | GCP credentials                     |
| **Best For**    | macOS dev/test          | Linux dev/test                | Premium quality                       | Enterprise/Scale                    |

## Detailed Comparison

### Voice Quality

#### say (macOS)

- **Pros**: Natural-sounding, good for local testing
- **Cons**: Not neural TTS, somewhat robotic
- **Use Case**: Development, testing, local projects

#### espeak (Linux)

- **Pros**: Lightweight, fast, open source
- **Cons**: Robotic voice, limited expressiveness
- **Use Case**: Development, testing, scripting

#### ElevenLabs

- **Pros**: Highly realistic, emotional control, voice cloning
- **Cons**: Requires paid subscription, limited speed range
- **Use Case**: Production content, audiobooks, podcasts

#### Google Cloud TTS

- **Pros**: Neural2/WaveNet voices, massive voice library, enterprise SLA
- **Cons**: Requires GCP setup, costs scale with usage
- **Use Case**: Enterprise, multi-language, high-volume

### Speed Control Comparison

| Provider     | Speed Range | Timing Accuracy | Notes                           |
| ------------ | ----------- | --------------- | ------------------------------- |
| say          | 90-360 WPM  | ±1-3 seconds    | Wide range, good flexibility    |
| espeak       | Variable    | ±2-4 seconds    | Adjusts rate parameter          |
| ElevenLabs   | 0.7x-1.2x   | ±15%            | Limited range, natural quality  |
| Google Cloud | 0.25x-4.0x  | ±10%            | **Widest range**, best accuracy |

**For precise timing control**: Google Cloud TTS (0.25x-4.0x range)
**For natural quality**: ElevenLabs (limited but realistic)
**For flexibility**: say (wide WPM range)

### Voice Selection

#### say (macOS)

- ~70 voices across 30+ languages
- Organized by language/region
- Good variety, standard quality
- List with: `./md2audio -list-voices`

#### espeak (Linux)

- ~50 voices across 50+ languages
- Simple language codes (en-us, en-gb, etc.)
- Open source voice synthesis
- List with: `./md2audio -provider espeak -list-voices`

#### ElevenLabs

- 20+ professional voices
- Highly distinctive personalities
- Voice cloning available (paid tiers)
- Emotional range control
- List with: `./md2audio -provider elevenlabs -list-voices`

#### Google Cloud TTS

- **400+ voices** across 50+ languages
- Multiple quality tiers:
  - Standard (basic quality)
  - WaveNet (high quality)
  - Neural2 (best quality)
  - Studio (premium, highest fidelity)
  - Polyglot (multi-language)
- List with: `./md2audio -provider google -list-voices`

### Output Formats

| Provider     | Formats             | Notes                          |
| ------------ | ------------------- | ------------------------------ |
| say          | AIFF, M4A           | AIFF default, converts to M4A  |
| espeak       | WAV, MP3, M4A, AIFF | WAV default, ffmpeg for others |
| ElevenLabs   | MP3                 | MP3 only from API              |
| Google Cloud | MP3, WAV, OGG       | Multiple formats supported     |

## Next Steps

- [say Provider Guide](providers/say.md)
- [espeak Provider Guide](providers/espeak.md)
- [ElevenLabs Provider Guide](providers/elevenlabs.md)
- [Google Cloud TTS Provider Guide](providers/google.md)
- [Timing Control Guide](timing-guide.md)
