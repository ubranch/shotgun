# Development Commands

## Setup
```bash
# Install dependencies
cd frontend && pnpm install

# Set Google API key for Gemini token counting
# Windows (PowerShell):
$env:GOOGLE_API_KEY = "your-api-key"

# Unix-like systems:
export GOOGLE_API_KEY="your-api-key"
```

## Development
```bash
# Run in development mode
wails dev

# Build for production
wails build
```

## System Commands (Windows)
- List files: `dir`
- Remove file: `del file.txt`
- Remove directory: `rmdir /s /q dir`
- Copy file: `copy source.txt destination.txt`
- Create directory: `mkdir dir`
- View file content: `type file.txt`
- Command separator: `&`