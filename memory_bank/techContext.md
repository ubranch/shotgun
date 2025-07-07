# technical context

## tech stack

- **backend**: go 1.24.0+ with wails framework
- **frontend**: vue.js, tailwind css, vite
- **apis**: google ai api for gemini token counting
- **file system**: native go file system operations
- **clipboard**: wails clipboard integration

## development environment

- **platform**: windows (primarily), cross-platform support
- **build system**: wails build system
- **package management**: pnpm for frontend, go modules for backend
- **environment variables**: google api key for gemini features

## technical constraints

- requires valid google api key for accurate token counting
- falls back to error message if api key is not available
- designed for desktop use (wails application)
- requires go 1.24.0+ and node.js for development
