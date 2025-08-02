# Technology Stack

## Backend
- **Go 1.24.0+** - Main application backend
- **Wails v2** - Desktop application framework for Go + frontend
- **Google Generative AI Go SDK** - For Gemini token counting
- **fsnotify** - File system watching
- **godirwalk** - Efficient directory traversal
- **go-gitignore** - Gitignore pattern matching

## Frontend
- **Vue.js 3** - Frontend framework
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Utility-first CSS framework
- **VS Code Codicons** - Icon library
- **pnpm** - Package manager (preferred over npm/yarn)

## Build System & Commands

### Development
```bash
# Start development mode (hot reload)
wails dev
```

### Production Build
```bash
# Build for production
wails build
```

### Frontend Development
```bash
# Install frontend dependencies
cd frontend && pnpm install

# Run frontend dev server (standalone)
cd frontend && pnpm run dev

# Build frontend only
cd frontend && pnpm run build
```

## Environment Setup

### Required Environment Variables
- `GOOGLE_API_KEY` - Google AI API key for Gemini token counting

### Platform-specific Setup
**Windows (PowerShell):**
```powershell
$env:GOOGLE_API_KEY = "your-api-key"
```

**Unix-like systems:**
```bash
export GOOGLE_API_KEY="your-api-key"
```

## Key Dependencies
- **Wails configuration** in `wails.json`
- **Go modules** managed via `go.mod`
- **Frontend deps** in `frontend/package.json`
- **Tailwind config** with CSS variables for theming
- **Vite config** with Vue plugin
