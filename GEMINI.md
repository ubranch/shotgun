# Gemini Project: shotgun

## Project Overview

**shotgun** is a desktop application designed for streamlined prompt engineering and testing. It provides a user-friendly interface for crafting prompts, managing context from local files, and executing these prompts with AI models. The application is built using a combination of Go for the backend and a Vue.js single-page application for the frontend, all packaged into a desktop app using Wails.

### Core Technologies:

*   **Backend:** Go
*   **Frontend:** Vue.js
*   **Desktop App Framework:** Wails
*   **Styling:** Tailwind CSS
*   **AI Integration:** Google AI SDK for token counting

### Architecture:

The application follows a typical Wails architecture:

*   A Go backend (`main.go`, `app.go`) handles the core application logic, including file system interactions, token counting, and communication with the frontend.
*   A Vue.js frontend (`frontend/` directory) provides the user interface. The root component is `App.vue`, which uses a `MainLayout` to structure the UI.
*   Wails acts as a bridge between the Go backend and the Vue.js frontend, allowing them to communicate with each other.

## Building and Running

### Prerequisites:

*   Go 1.24.0 or higher
*   Node.js and pnpm
*   A Google API key for Gemini token counting, set as the `GOOGLE_API_KEY` environment variable.

### Development Mode:

To run the application in development mode with live reloading, use the following command from the project root:

```bash
wails dev
```

### Production Build:

To build a production version of the application, use the following command from the project root:

```bash
wails build
```

This will create a native executable for your operating system in the `build/bin` directory.

## Development Conventions

### Backend (Go):

*   The main application logic is encapsulated in the `App` struct in `app.go`.
*   The `main.go` file is responsible for initializing the Wails application, setting up the window, and binding the `App` struct to the frontend.
*   Dependencies are managed using Go modules (`go.mod` and `go.sum`).

### Frontend (Vue.js):

*   The frontend source code is located in the `frontend/src` directory.
*   The application uses Vue 3 with the Composition API (`<script setup>`).
*   Components are organized in the `frontend/src/components` directory.
*   Styling is done using Tailwind CSS, with the configuration in `tailwind.config.js`.
*   Frontend dependencies are managed using pnpm (`package.json` and `pnpm-lock.yaml`).
*   The frontend is built using Vite (`vite.config.js`).
