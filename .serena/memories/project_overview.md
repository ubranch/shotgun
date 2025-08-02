# Shotgun - Project Overview

## Purpose
Shotgun is a tool for streamlined prompt engineering and testing. It provides a user-friendly interface for crafting prompts, managing context, and executing prompts with AI models.

## Tech Stack
- **Backend**: Go with Wails v2 framework
- **Frontend**: Vue.js 3 with Composition API
- **Styling**: Tailwind CSS
- **Build Tool**: Vite
- **Package Manager**: pnpm

## Key Features
- Accurate Gemini token counting using the official Google AI SDK
- File tree visualization with gitignore support
- Customizable prompt templates
- Live token counting and validation
- Clipboard integration
- Context generation from project files
- Multi-step workflow for prompt engineering

## Architecture
- Wails app with Go backend and Vue frontend
- Event-driven communication between frontend and backend
- File watching and context generation
- Debounced operations for performance