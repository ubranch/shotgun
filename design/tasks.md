# Tasks Log

## Task 1: Display Project Root Folder in Tree

*   **Date Completed:** (Fill this in with today's date)
*   **Problem:** The file tree displayed the contents of the selected project directory as top-level items. The selected directory itself was not shown as a node, making it difficult to exclude all items or manage files directly in the root via the tree UI.
*   **Solution:**
    *   **Backend (`app.go`):**
        *   Modified `ListFiles(dirPath string)` to create a single root `FileNode` representing `dirPath`.
        *   The `Name` of this root node is `filepath.Base(dirPath)`.
        *   Its `RelPath` is set to `"."`.
        *   `IsDir` is `true`.
        *   `IsGitignored` is `false` (by default for the root).
        *   The `Children` of this root node are populated by the original `buildTree` logic, which now processes the content of `dirPath`.
        *   `ListFiles` now returns `[]*FileNode` containing only this single root node.
    *   **Frontend (`App.vue`, `FileTree.vue`):**
        *   No significant changes were anticipated to be required. The existing logic in `App.vue` for `loadFileTree` and `mapDataToTree` should correctly handle an array with a single root node.
        *   The recursive `FileTree.vue` component should also correctly render the tree starting from a single root.
        *   **Improvement (YYYY-MM-DD):** Modified `mapDataToTree` in `App.vue` so that the root node (`parent === null`) has its `expanded` property set to `true` by default. This ensures the first level of the project tree is open when first loaded.
    *   **Documentation:**
        *   Updated `design/architecture.md` to reflect the new behavior of `ListFiles` and the resulting tree structure.
        *   Added this entry to `design/tasks.md` (and updated with this improvement).
*   **Status:** Backend changes implemented and documented. Frontend behavior confirmed by analysis to be compatible. Root node now expands by default.

# Project Tasks

## Phase 1: Initial ignore.glob Implementation (Completed)

- [x] Create `ignore.glob` with default media file patterns.
- [x] Modify `app.go` (`ListFiles`, `buildTree`) to parse `ignore.glob` and combine its rules with `.gitignore`.

## Phase 2: Configurable Ignore Patterns

- [ ] Вынести шаблоны из `ignore.glob` в настройки приложения, чтобы пользователь мог их редактировать через интерфейс приложения.
- [ ] Обеспечить сохранение пользовательских шаблонов между сессиями приложения.

- [x] Add platform detection (using Wails Environment API) in MainLayout.vue and pass it down to CentralPanel and step components.
- [x] In Step1PrepareContext.vue and Step2ComposePrompt.vue, use WailsClipboardSetText for macOS (darwin), otherwise use navigator.clipboard for copying to clipboard.
- [x] Update CentralPanel.vue and MainLayout.vue to forward platform prop.
- [x] Update prop definitions and usages in all affected components.
- [x] Update Go main.go to use os.ReadFile and add menu for macOS.
