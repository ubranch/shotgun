# shotgun - project structure

## description

shotgun is a desktop application built with go (backend) and vue.js (frontend) using the wails framework. it helps developers by generating a comprehensive "snapshot" of their codebase that can be fed into large language models (llms) like chatgpt, gemini, or cursor. this allows for multi-file edits, refactors, bug fixes, and code reviews with proper context.

## core functionality

-   scanning and representing file trees
-   generating a structured text representation of project files
-   custom ignore rules (similar to gitignore)
-   file watching for changes
-   splitting large diffs into manageable chunks
-   4-step workflow for context preparation, prompt composition, execution, and patch application

## root files

-   `app.go` - main application logic including context generation, file watching, and settings management
-   `main.go` - entry point for the wails application setup
-   `split_diff.go` - functionality for splitting large git diffs into smaller chunks
-   `go.mod` - go module definition for shotgun
-   `go.sum` - go dependencies checksums
-   `wails.json` - wails configuration file
-   `readme.md` - project documentation
-   `license.md` - license information
-   `ignore.glob` - default custom ignore rules for file filtering
-   `appicon.ico`, `appicon.icns`, `appicon.png` - application icons for different platforms

## directory structure

### frontend

-   `frontend/index.html` - html entry point for the vue application
-   `frontend/package.json` - npm package definition
-   `frontend/tailwind.config.js` - tailwind css configuration
-   `frontend/vite.config.js` - vite configuration

### frontend/src

-   `frontend/src/main.js` - javascript entry point for the vue application
-   `frontend/src/app.vue` - main vue application component with theme provider setup

### frontend/src/assets

-   `frontend/src/assets/main.css` - main css styles
-   `frontend/src/assets/custom.css` - custom css styles with font definitions
-   `frontend/src/assets/fonts/` - directory containing font files for the application

### frontend/src/components

-   `frontend/src/components/mainlayout.vue` - main layout component organizing the ui structure
-   `frontend/src/components/horizontalstepper.vue` - horizontal step visualization in top bar with consistent font weight to prevent layout shifts
-   `frontend/src/components/centralpanel.vue` - central panel component for displaying step content
-   `frontend/src/components/filetree.vue` - file tree component for visualizing project structure
-   `frontend/src/components/leftsidebar.vue` - left sidebar component for project selection and file tree
-   `frontend/src/components/bottomconsole.vue` - bottom console component for logs and messages
-   `frontend/src/components/customrulesmodal.vue` - modal for configuring custom ignore rules
-   `frontend/src/components/themeprovider.vue` - dark/light theme provider component
-   `frontend/src/components/themetoggle.vue` - theme toggle button component

### frontend/src/components/steps

-   `frontend/src/components/steps/step1preparecontext.vue` - step 1 view for preparing project context
-   `frontend/src/components/steps/step2composeprompt.vue` - step 2 view for composing prompts to llms
-   `frontend/src/components/steps/step3executeprompt.vue` - step 3 view for executing prompts
-   `frontend/src/components/steps/step4applypatch.vue` - step 4 view for applying generated patches

### design

-   `design/prompts/` - directory containing prompt templates for different use cases

### wailsjs

-   `frontend/wailsjs/go/main/app.js` - javascript bindings for go functions
-   `frontend/wailsjs/runtime/runtime.js` - wails runtime javascript utilities
