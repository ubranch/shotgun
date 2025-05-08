![](https://github.com/user-attachments/assets/ec9c0d87-fc7c-4800-a52e-a34356ef09c3)

# Shotgun App

*One‑click codebase "blast" for Large‑Language‑Model workflows.*

---

## 1. What Shotgun Does
Shotgun is a tiny desktop tool that **explodes an entire project into a single,
well‑structured text payload** designed for AI assistants.
Think of it as a rapid‑fire alternative to copy‑pasting dozens of files by hand:

*   **Select a folder → get an instant tree + file dump**
    in a predictable delimiter format (`*#*#*...*#*#*begin … *#*#*end*#*#*`).
*   **Tick check‑boxes to exclude noise** (logs, build artifacts, `node_modules`, …).
*   **Paste the result into ChatGPT, Gemini 2.5, Cursor, etc.**
    to ask for multi‑file edits, refactors, bug fixes, reviews, or documentation.
*   **Receive a diff‑style reply** and apply changes with your favourite patch tool.

Shotgun trades surgical, single‑file prompts for a **"whole‑repository blast"** –
hence the name.

---

## 2. Why You Might Need It

| Scenario                 | Pain Point                             | Shotgun Benefit                                           |
|--------------------------|----------------------------------------|-----------------------------------------------------------|
| **Bulk bug fixing**      | "Please fix X across 12 files."        | Generates a complete snapshot so the LLM sees all usages. |
| **Large‑scale refactor** | IDE refactors miss edge cases.         | LLM gets full context and returns a patch set.            |
| **On‑boarding review**   | New joiner must understand legacy code. | Produce a single, searchable text file to discuss in chat.  |
| **Doc generation**       | Want docs/tests for every exported symbol. | LLM can iterate over full source without extra API calls. |
| **Cursor / CodePilot prompts** | Tools accept pasted context but no filesystem. | Shotgun bridges the gap.                                  |

---

## 3. Key Features

*   ⚡ **Fast tree scan** (Go + Wails backend) – thousands of files in milliseconds.
*   ✅ **Interactive exclude list** – skip folders, temporary files, or secrets.
*   📝 **Deterministic delimiters** – easy for LLMs to parse and for you to split.
*   🔄 **Re‑generate anytime** – tweak the excludes and hit *Shotgun* again.
*   🪶 **Lightweight** – no DB, no cloud; a single native executable plus a Vue UI.
*   🖥️ **Cross‑platform** – Windows, macOS, Linux.

---

## 4. How It Works

1.  **Select project root**
    Shotgun uses Wails' native file‑picker to choose a directory.
2.  **Build an in‑memory tree** (`ListFiles` in `app.go`).
    Directories are listed first, then files, both sorted alphabetically.
3.  **Mark exclusions** in the tree component (Vue 3).
    Check a node → all children grey out.
4.  **Click "Shotgun!"**
    `GenerateShotgunOutput` walks the tree, skipping excluded paths, and
    concatenates:
    ```text
    project‑tree.txt

    ##src/main.go##begin##*
    // file contents …
    ##end##

    ##web/index.html##begin##*

    <!-- file contents … -->
    ##end##
    ```
5.  **Paste into an LLM prompt**, for example:

    > I have a race condition (details…).
    > Here is the entire codebase from Shotgun. Please fix it and return diffs
    > in the following format …

6.  **Apply the diff** the model returns.

---

## 5. Installation

### 5.1. Prerequisites
*   **Go ≥ 1.20**   `go version`
*   **Node.js LTS**  `node -v`
*   **Wails CLI**    `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 5.2. Clone & Bootstrap
```bash
git clone https://github.com/your‑org/shotgun‑app.git
cd shotgun‑app
go mod tidy           # backend deps
cd frontend
npm install           # Vue deps
cd ..
```

### 5.3. Run in Dev Mode
```bash
wails dev
```
Hot‑reloads Vue; restart the command for Go code changes.

### 5.4. Build a Release
```bash
wails build           # binaries land in build/bin/
```

---

## 6. Quick‑Start Workflow

1.  Run `wails dev`. The app window will open.
2.  Click "Select Project Folder" and choose your repository root.
3.  In the left pane, expand folders and un-tick any items you wish to exclude.
4.  Hit the "Shotgun" button. The right pane will now show the formatted dump.
5.  Use the "Copy to Clipboard" button (or select all with Ctrl/Cmd + A, then copy with Ctrl/Cmd + C).
6.  Paste the copied content into your AI assistant along with your instructions. For example:
    ```text
    ### Prompt
    I need to migrate this project from Go 1.20 to 1.22.
    - Update go.mod, replace deprecated APIs.
    - Return patches in the exact diff block format below.

    [paste Shotgun output here]
    ```
7.  Apply the diffs provided by the assistant using `git apply` or a merge tool.

---

## 7. Shotgun Output Anatomy
```text
app/
├── main.go
├── app.go
└── frontend/
    ├── App.vue
    └── components/
        └── FileTree.vue

*#*#*main.go*#*#*begin*#*#*
package main
...
*#*#*end*#*#*

*#*#*frontend/components/FileTree.vue*#*#*begin*#*#*
<template>
...
*#*#*end*#*#*
```
*   **Tree View** – quick visual map for you & the LLM.
*   **Delimited File Blocks** – deterministic markers so models can chunk
    input or generate per‑file diffs (`*#*#*path*#*#*begin` / `end*#*#*`).

---

## 8. Best Practices
*   **Trim the noise** – exclude lock files, vendored libs, generated assets.
    Less tokens → cheaper & more accurate completions.
*   **Ask for diffs, not whole files** – keeps responses concise.
*   **Iterate** – generate → ask → patch → re‑generate if needed.
*   **Watch token limits** – even million‑token models have practical caps.
    Use Shotgun scopes (root folder vs subfolder) to stay under budget.

---

## 9. Troubleshooting

| Symptom                     | Fix                                                          |
|-----------------------------|--------------------------------------------------------------|
| `wails: command not found`  | Ensure `$GOROOT/bin` or `$HOME/go/bin` is on `PATH`.         |
| Blank window on `wails dev` | Check Node version & reinstall frontend deps.              |
| Output too large            | Split Shotgun runs by subdirectory; or exclude binaries/tests. |

---

## 10. Roadmap
*   🗂️ Preset exclude templates (Go, Node, Rust, …).
*   🧠 Direct API bridge to send output to OpenAI / Gemini without copy‑paste.
*   🔀 Patch importer – apply AI‑generated diffs from inside Shotgun.
*   📦 CLI version for headless pipelines.

---

## 11. Contributing
PRs and issues are welcome!
Please format Go code with `go fmt` and follow Vue 3 style guidelines.

---

## 12. License
MIT – see `LICENSE` file.

---

Shotgun – load, aim, blast your code straight into the mind of an LLM.
Iterate faster. Ship better. 
