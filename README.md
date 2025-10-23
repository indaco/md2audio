<h1 align="center">
    md2audio
</h1>
<h2 align="center" style="font-size: 1.5em;">
  Markdown to Audio Generator.
</h2>
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

Convert markdown H2 sections to individual audio files using macOS `say` command.

## Features

- Process single files or entire directories recursively
- Automatically extracts H2 sections from markdown files
- Generates separate audio file for each section
- Mirror directory structure for batch processing
- Configurable voice (British, US, Australian, etc.)
- Adjustable speaking rate for clarity
- Supports AIFF and M4A output formats
- Clean filename generation from section titles
- Target duration control with timing annotations

## Prerequisites

- macOS (uses `say` command)
- Go 1.25 or later (to build the tool)

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

## Usage

### Basic Examples

```bash
# List available voices
./md2audio -list-voices

# Process a single markdown file
./md2audio -f script.md -p british-female

# Process entire directory recursively
./md2audio -d ./docs -p british-female

# Use specific voice with slower rate for clarity
./md2audio -f script.md -v Kate -r 170

# Generate M4A files instead of AIFF
./md2audio -d ./content -p british-female -format m4a

# Custom output directory and prefix
./md2audio -f script.md -o ./voiceovers -prefix demo
```

### Command Line Options

| Flag           | Description                                   | Default             |
| -------------- | --------------------------------------------- | ------------------- |
| `-f`           | Input markdown file (use `-f` or `-d`)        | -                   |
| `-d`           | Input directory (recursive, use `-f` or `-d`) | -                   |
| `-o`           | Output directory                              | `./audio_sections`  |
| `-p`           | Voice preset                                  | `Kate` (if not set) |
| `-v`           | Specific voice name (overrides preset)        | -                   |
| `-r`           | Speaking rate (lower = slower)                | `180`               |
| `-format`      | Output format (`aiff` or `m4a`)               | `aiff`              |
| `-prefix`      | Filename prefix                               | `section`           |
| `-list-voices` | List all available voices                     | -                   |

### Voice Presets

- `british-female` -> Kate
- `british-male` -> Daniel
- `us-female` -> Samantha
- `us-male` -> Alex
- `australian-female` -> Karen
- `indian-female` -> Veena

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

**How it works:**

- The script counts the words in your text
- Calculates the required words-per-minute (WPM) to fit the target duration
- Automatically adjusts the speaking rate for that section
- Shows you the actual duration vs target after generation

**Note on timing accuracy:** Due to limitations in macOS's `say` command rate handling, the actual duration may differ from the target (typically within 1-3 seconds). The tool applies a 0.95 adjustment factor to improve accuracy, but exact matches are not guaranteed. Test with your content and adjust timing annotations as needed.

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

The codebase is organized into modular packages for better maintainability and testability:

```
md2audio/
├── cmd/md2audio/        # Main entry point (orchestration only)
│   ├── main.go
├── internal/
│   ├── config/          # Configuration and CLI flags
│   ├── parser/          # Markdown parsing and file discovery
│   ├── text/            # Text processing utilities
│   ├── audio/           # Audio generation logic
└── └── processor/       # File and directory processing (NEW)
```

### Key Packages

- **internal/config** - Handles command-line arguments, voice presets, and configuration validation
- **internal/parser** - Extracts H2 sections from markdown with timing annotations, discovers markdown files recursively
- **internal/text** - Provides markdown cleaning and filename sanitization
- **internal/audio** - Core audio generation, rate calculation, and format conversion
- **internal/processor** - Orchestrates file and directory processing with mirror structure support

## Contributing

Contributions are welcome!

See the [Contributing Guide](/CONTRIBUTING.md) for setup instructions.

## License

This project is licensed under the MIT License – see the [LICENSE](./LICENSE) file for details.
