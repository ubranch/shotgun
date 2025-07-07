package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"bytes"
	"unicode/utf8"

	"github.com/adrg/xdg"
	"github.com/fsnotify/fsnotify"
	"github.com/google/generative-ai-go/genai"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"google.golang.org/api/option"
)

const maxOutputSizeBytes = 10_000_000 // 10mb
var ErrContextTooLong = errors.New("context is too long")

//go:embed ignore.glob
var defaultCustomIgnoreRulesContent string

const defaultCustomPromptRulesContent = "no additional rules"

type AppSettings struct {
	CustomIgnoreRules string `json:"customIgnoreRules"`
	CustomPromptRules string `json:"customPromptRules"`
}

type App struct {
	ctx                         context.Context
	contextGenerator            *ContextGenerator
	fileWatcher                 *Watchman
	settings                    AppSettings
	currentCustomIgnorePatterns *gitignore.GitIgnore
	configPath                  string
	useGitignore                bool
	useCustomIgnore             bool
	projectGitignore            *gitignore.GitIgnore // compiled .gitignore for the current project
	geminiRequestCancel         context.CancelFunc   // cancel function for gemini request

	// defaultrootdir holds an optional folder path passed via command line argument (e.g. when a user
	// drags a folder onto the compiled executable). if set, the app will emit an event on startup so
	// it can open the folder automatically.
	defaultRootDir string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.contextGenerator = NewContextGenerator(a)
	a.fileWatcher = NewWatchman(a)
	a.useGitignore = true    // default to true, matching frontend
	a.useCustomIgnore = true // default to true, matching frontend

	configFilePath, err := xdg.ConfigFile("shotgun-code/settings.json")
	if err != nil {
		runtime.LogErrorf(a.ctx, "error getting config file path: %v. using defaults and will attempt to save later if rules are modified.", err)
		// configpath will be empty, loadsettings will handle this by using defaults
		// and savesettings will fail gracefully if configpath remains empty and saving is attempted.
	}
	a.configPath = configFilePath

	a.loadSettings()
	// ensure custompromptrules has a default if it's empty after loading
	if strings.TrimSpace(a.settings.CustomPromptRules) == "" {
		a.settings.CustomPromptRules = defaultCustomPromptRulesContent
	}

	// if a default root directory was provided we will emit an auto-open event
	// after the frontend is fully ready (see domready). here we just set the
	// window title early for better ux.
	if a.defaultRootDir != "" {
		if info, err := os.Stat(a.defaultRootDir); err == nil && info.IsDir() {
			folderName := filepath.Base(a.defaultRootDir)
			title := fmt.Sprintf("%s | Shotgun", folderName)
			runtime.WindowSetTitle(a.ctx, title)
		} else {
			runtime.LogWarningf(a.ctx, "startup: provided defaultRootDir '%s' is invalid: %v", a.defaultRootDir, err)
			// invalidate if not valid
			a.defaultRootDir = ""
		}
	}
}

// domready is called by wails when the frontend has finished loading and the js
// runtime is ready. only at this point are event listeners on the js side able
// to receive events, so we emit the auto-open-folder event here.
func (a *App) domReady(ctx context.Context) {
	if a.defaultRootDir == "" {
		return
	}
	runtime.EventsEmit(ctx, "auto-open-folder", a.defaultRootDir)
}

type FileNode struct {
	Name            string      `json:"name"`
	Path            string      `json:"path"`    // full path
	RelPath         string      `json:"relPath"` // path relative to selected root
	IsDir           bool        `json:"isDir"`
	Children        []*FileNode `json:"children,omitempty"`
	IsGitignored    bool        `json:"isGitignored"`    // true if path matches a .gitignore rule
	IsCustomIgnored bool        `json:"isCustomIgnored"` // true if path matches a ignore.glob rule
}

// selectdirectory opens a dialog to select a directory and returns the chosen path (empty string on cancel)
func (a *App) SelectDirectory() (string, error) {
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})

	// if the dialog was closed or cancelled, wails may return an error with an empty path
	// we treat this as a normal cancel action: return empty path and nil error so the
	// frontend does not show a scary error message.
	if err != nil && dirPath == "" {
		// optionally log at debug level for diagnostics
		runtime.LogDebugf(a.ctx, "selectdirectory: dialog closed without selection: %v", err)
		return "", nil
	}

	if err != nil {
		// genuine error where a path should have been returned; propagate it
		return "", err
	}

	if dirPath != "" {
		folderName := filepath.Base(dirPath)
		title := fmt.Sprintf("%s | Shotgun", folderName)
		runtime.WindowSetTitle(a.ctx, title)
	}
	return dirPath, nil
}

// listfiles lists files and folders in a directory, parsing .gitignore if present
func (a *App) ListFiles(dirPath string) ([]*FileNode, error) {
	runtime.LogDebugf(a.ctx, "listfiles called for directory: %s", dirPath)

	a.projectGitignore = nil        // reset for the new directory
	var gitIgn *gitignore.GitIgnore // for .gitignore in the project directory
	gitignorePath := filepath.Join(dirPath, ".gitignore")
	runtime.LogDebugf(a.ctx, "attempting to find .gitignore at: %s", gitignorePath)
	if _, err := os.Stat(gitignorePath); err == nil {
		runtime.LogDebugf(a.ctx, ".gitignore found at: %s", gitignorePath)
		gitIgn, err = gitignore.CompileIgnoreFile(gitignorePath)
		if err != nil {
			runtime.LogWarningf(a.ctx, "error compiling .gitignore file at %s: %v", gitignorePath, err)
			gitIgn = nil
		} else {
			a.projectGitignore = gitIgn // store the compiled project-specific gitignore
			runtime.LogDebug(a.ctx, ".gitignore compiled successfully.")
		}
	} else {
		runtime.LogDebugf(a.ctx, ".gitignore not found at %s (os.stat error: %v)", gitignorePath, err)
		gitIgn = nil
	}

	// app-level custom ignore patterns are in a.currentcustomignorepatterns

	rootNode := &FileNode{
		Name:         filepath.Base(dirPath),
		Path:         dirPath,
		RelPath:      ".",
		IsDir:        true,
		IsGitignored: false, // root itself is not gitignored by default
		// iscustomignored for root is also false by default, specific patterns would be needed
		IsCustomIgnored: a.currentCustomIgnorePatterns != nil && a.currentCustomIgnorePatterns.MatchesPath("."),
	}

	children, err := buildTreeRecursive(context.TODO(), dirPath, dirPath, gitIgn, a.currentCustomIgnorePatterns, 0)
	if err != nil {
		return []*FileNode{rootNode}, fmt.Errorf("error building children tree for %s: %w", dirPath, err)
	}
	rootNode.Children = children

	return []*FileNode{rootNode}, nil
}

func buildTreeRecursive(ctx context.Context, currentPath, rootPath string, gitIgn *gitignore.GitIgnore, customIgn *gitignore.GitIgnore, depth int) ([]*FileNode, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}

	var nodes []*FileNode
	for _, entry := range entries {
		nodePath := filepath.Join(currentPath, entry.Name())
		relPath, _ := filepath.Rel(rootPath, nodePath)
		// for gitignore matching, paths should generally be relative to the .gitignore file (rootpath)
		// and use os-specific separators. go-gitignore handles this.

		isGitignored := false
		isCustomIgnored := false
		pathToMatch := relPath
		if entry.IsDir() {
			if !strings.HasSuffix(pathToMatch, string(os.PathSeparator)) {
				pathToMatch += string(os.PathSeparator)
			}
		}

		if gitIgn != nil {
			isGitignored = gitIgn.MatchesPath(pathToMatch)
		}
		if customIgn != nil {
			isCustomIgnored = customIgn.MatchesPath(pathToMatch)
		}

		if depth < 2 || strings.Contains(relPath, "node_modules") || strings.HasSuffix(relPath, ".log") {
			fmt.Printf("checking path: '%s' (original relpath: '%s'), isdir: %v, gitignored: %v, customignored: %v\n", pathToMatch, relPath, entry.IsDir(), isGitignored, isCustomIgnored)
		}

		node := &FileNode{
			Name:            entry.Name(),
			Path:            nodePath,
			RelPath:         relPath,
			IsDir:           entry.IsDir(),
			IsGitignored:    isGitignored,
			IsCustomIgnored: isCustomIgnored,
		}

		if entry.IsDir() {
			// if it's a directory, recursively call buildtree
			// only recurse if not ignored
			if !isGitignored && !isCustomIgnored {
				children, err := buildTreeRecursive(ctx, nodePath, rootPath, gitIgn, customIgn, depth+1)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil, err // propagate cancellation
					}
					// runtime.logwarnf(ctx, "error building subtree for %s: %v", nodepath, err) // use ctx if available
					runtime.LogWarningf(context.Background(), "error building subtree for %s: %v", nodePath, err) // fallback for now
					// decide: skip this dir or return error up. for now, skip with log.
				} else {
					node.Children = children
				}
			}
		}
		nodes = append(nodes, node)
	}
	// sort nodes: directories first, then files, then alphabetically
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir && !nodes[j].IsDir {
			return true
		}
		if !nodes[i].IsDir && nodes[j].IsDir {
			return false
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
	return nodes, nil
}

// contextgenerator manages the asynchronous generation of shotgun context
type ContextGenerator struct {
	app                *App // to access wails runtime context for emitting events
	mu                 sync.Mutex
	currentCancelFunc  context.CancelFunc
	currentCancelToken interface{} // token to identify the current cancel func
}

func NewContextGenerator(app *App) *ContextGenerator {
	return &ContextGenerator{app: app}
}

// requestshotguncontextgeneration is called by the frontend to start/restart generation.
// this method itself is not bound to wails directly if it's part of app.
// instead, a wrapper method in app struct will be bound.
func (cg *ContextGenerator) requestShotgunContextGenerationInternal(rootDir string, excludedPaths []string) {
	cg.mu.Lock()
	if cg.currentCancelFunc != nil {
		runtime.LogDebug(cg.app.ctx, "cancelling previous context generation job.")
		cg.currentCancelFunc()
	}

	genCtx, cancel := context.WithCancel(cg.app.ctx)
	myToken := new(struct{}) // create a unique token for this generation job
	cg.currentCancelFunc = cancel
	cg.currentCancelToken = myToken
	runtime.LogInfof(cg.app.ctx, "starting new shotgun context generation for: %s. max size: %d bytes.", rootDir, maxOutputSizeBytes)
	cg.mu.Unlock()

	go func(tokenForThisJob interface{}) {
		jobStartTime := time.Now()
		defer func() {
			cg.mu.Lock()
			if cg.currentCancelToken == tokenForThisJob { // only clear if it's still this job's token
				cg.currentCancelFunc = nil
				cg.currentCancelToken = nil
				runtime.LogDebug(cg.app.ctx, "cleared currentcancelfunc for completed/cancelled job (token match).")
			} else {
				runtime.LogDebug(cg.app.ctx, "currentcancelfunc was replaced by a newer job (token mismatch); not clearing.")
			}
			cg.mu.Unlock()
			runtime.LogInfof(cg.app.ctx, "shotgun context generation goroutine finished in %s", time.Since(jobStartTime))
		}()

		if genCtx.Err() != nil { // check for immediate cancellation
			runtime.LogInfo(cg.app.ctx, fmt.Sprintf("context generation for %s cancelled before starting: %v", rootDir, genCtx.Err()))
			return
		}

		output, err := cg.app.generateShotgunOutputWithProgress(genCtx, rootDir, excludedPaths)

		select {
		case <-genCtx.Done():
			errMsg := fmt.Sprintf("shotgun context generation cancelled for %s: %v", rootDir, genCtx.Err())
			runtime.LogInfo(cg.app.ctx, errMsg) // changed from logwarn
			runtime.EventsEmit(cg.app.ctx, "shotgunContextError", errMsg)
		default:
			if err != nil {
				errMsg := fmt.Sprintf("error generating shotgun output for %s: %v", rootDir, err)
				runtime.LogError(cg.app.ctx, errMsg)
				runtime.EventsEmit(cg.app.ctx, "shotgunContextError", errMsg)
			} else {
				finalSize := len(output)
				successMsg := fmt.Sprintf("shotgun context generated successfully for %s. size: %d bytes.", rootDir, finalSize)
				if finalSize > maxOutputSizeBytes { // should have been caught by errcontexttoolong, but as a safeguard
					runtime.LogWarningf(cg.app.ctx, "warning: generated context size %d exceeds max %d, but was not caught by errcontexttoolong.", finalSize, maxOutputSizeBytes)
				}
				runtime.LogInfo(cg.app.ctx, successMsg)
				runtime.EventsEmit(cg.app.ctx, "shotgunContextGenerated", output)
			}
		}
	}(myToken) // pass the token to the goroutine
}

// requestshotguncontextgeneration is the method bound to wails.
func (a *App) RequestShotgunContextGeneration(rootDir string, excludedPaths []string) {
	if a.contextGenerator == nil {
		// this should not happen if startup initializes it correctly
		runtime.LogError(a.ctx, "contextgenerator not initialized")
		runtime.EventsEmit(a.ctx, "shotgunContextError", "internal error: contextgenerator not initialized")
		return
	}
	a.contextGenerator.requestShotgunContextGenerationInternal(rootDir, excludedPaths)
}

// countprocessableitems estimates the total number of operations for progress tracking.
// operations: 1 for root dir line, 1 for each dir/file entry in tree, 1 for each file content read.
func (a *App) countProcessableItems(jobCtx context.Context, rootDir string, excludedMap map[string]bool) (int, error) {
	count := 1 // for the root directory line itself

	var counterHelper func(currentPath string) error
	counterHelper = func(currentPath string) error {
		select {
		case <-jobCtx.Done():
			return jobCtx.Err()
		default:
		}

		entries, err := os.ReadDir(currentPath)
		if err != nil {
			runtime.LogWarningf(a.ctx, "countprocessableitems: error reading dir %s: %v", currentPath, err)
			return nil // continue counting other parts if a subdir is inaccessible
		}

		for _, entry := range entries {
			path := filepath.Join(currentPath, entry.Name())
			relPath, _ := filepath.Rel(rootDir, path)

			if excludedMap[relPath] {
				continue
			}

			count++ // for the tree entry (dir or file)

			if entry.IsDir() {
				err := counterHelper(path)
				if err != nil { // propagate cancellation or critical errors
					return err
				}
			} else {
				count++ // for reading the file content
			}
		}
		return nil
	}

	err := counterHelper(rootDir)
	if err != nil {
		return 0, err // return error if counting was interrupted (e.g. context cancelled)
	}
	return count, nil
}

type generationProgressState struct {
	processedItems int
	totalItems     int
}

func (a *App) emitProgress(state *generationProgressState) {
	runtime.EventsEmit(a.ctx, "shotgunContextGenerationProgress", map[string]int{
		"current": state.processedItems,
		"total":   state.totalItems,
	})
}

// generateshotgunoutputwithprogress generates the txt output with progress reporting and size limits
func (a *App) generateShotgunOutputWithProgress(jobCtx context.Context, rootDir string, excludedPaths []string) (string, error) {
	if err := jobCtx.Err(); err != nil { // check for cancellation at the beginning
		return "", err
	}

	excludedMap := make(map[string]bool)
	for _, p := range excludedPaths {
		excludedMap[p] = true
	}

	totalItems, err := a.countProcessableItems(jobCtx, rootDir, excludedMap)
	if err != nil {
		return "", fmt.Errorf("failed to count processable items: %w", err)
	}
	progressState := &generationProgressState{processedItems: 0, totalItems: totalItems}
	a.emitProgress(progressState) // initial progress (0 / total)

	var output strings.Builder
	var fileContents strings.Builder

	// root directory line
	output.WriteString(filepath.Base(rootDir) + string(os.PathSeparator) + "\n")
	progressState.processedItems++
	a.emitProgress(progressState)
	if output.Len() > maxOutputSizeBytes {
		return "", fmt.Errorf("%w: content limit of %d bytes exceeded after root dir line (size: %d bytes)", ErrContextTooLong, maxOutputSizeBytes, output.Len())
	}

	// buildshotguntreerecursive is a recursive helper for generating the tree string and file contents
	var buildShotgunTreeRecursive func(pCtx context.Context, currentPath, prefix string) error
	buildShotgunTreeRecursive = func(pCtx context.Context, currentPath, prefix string) error {
		select {
		case <-pCtx.Done():
			return pCtx.Err()
		default:
		}

		entries, err := os.ReadDir(currentPath)
		if err != nil {
			runtime.LogWarningf(a.ctx, "buildshotguntreerecursive: error reading dir %s: %v", currentPath, err)
			// decide if this error should halt the entire process or just skip this directory
			// for now, returning nil to skip, but log it. could also return the error.
			return nil // or return err if this should stop everything
		}

		// sort entries like in listfiles for consistent tree
		sort.SliceStable(entries, func(i, j int) bool {
			entryI := entries[i]
			entryJ := entries[j]
			isDirI := entryI.IsDir()
			isDirJ := entryJ.IsDir()
			if isDirI && !isDirJ {
				return true
			}
			if !isDirI && isDirJ {
				return false
			}
			return strings.ToLower(entryI.Name()) < strings.ToLower(entryJ.Name())
		})

		// create a temporary slice to hold non-excluded entries for correct prefixing
		var visibleEntries []fs.DirEntry
		for _, entry := range entries {
			path := filepath.Join(currentPath, entry.Name())
			relPath, _ := filepath.Rel(rootDir, path)
			if !excludedMap[relPath] {
				visibleEntries = append(visibleEntries, entry)
			}
		}

		for i, entry := range visibleEntries {
			select {
			case <-pCtx.Done():
				return pCtx.Err()
			default:
			}

			path := filepath.Join(currentPath, entry.Name())
			relPath, _ := filepath.Rel(rootDir, path)

			isLast := i == len(visibleEntries)-1

			branch := "├── "
			nextPrefix := prefix + "│   "
			if isLast {
				branch = "└── "
				nextPrefix = prefix + "    "
			}
			output.WriteString(prefix + branch + entry.Name() + "\n")

			progressState.processedItems++ // for tree entry
			a.emitProgress(progressState)

			if output.Len()+fileContents.Len() > maxOutputSizeBytes {
				return fmt.Errorf("%w: content limit of %d bytes exceeded during tree generation (size: %d bytes)", ErrContextTooLong, maxOutputSizeBytes, output.Len()+fileContents.Len())
			}

			if entry.IsDir() {
				err := buildShotgunTreeRecursive(pCtx, path, nextPrefix)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						return err
					}
					fmt.Printf("error processing subdirectory %s: %v\n", path, err)
				}
			} else {
				select { // check before heavy i/o
				case <-pCtx.Done():
					return pCtx.Err()
				default:
				}
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("error reading file %s: %v\n", path, err)
					content = []byte(fmt.Sprintf("error reading file: %v", err))
				}

				// ensure forward slashes for the name attribute, consistent with documentation.
				relPathForwardSlash := filepath.ToSlash(relPath)

				fileContents.WriteString(fmt.Sprintf("<file path=\"%s\">\n", relPathForwardSlash))
				if isTextContent(content) {
					fileContents.WriteString(string(content))
				} else {
					fileContents.WriteString("[non-text file content omitted]")
				}
				fileContents.WriteString("\n</file>\n") // each file block ends with a newline

				progressState.processedItems++ // for file content
				a.emitProgress(progressState)

				if output.Len()+fileContents.Len() > maxOutputSizeBytes { // final check after append
					return fmt.Errorf("%w: content limit of %d bytes exceeded after appending file %s (total size: %d bytes)", ErrContextTooLong, maxOutputSizeBytes, relPath, output.Len()+fileContents.Len())
				}
			}
		}
		return nil
	}

	err = buildShotgunTreeRecursive(jobCtx, rootDir, "")
	if err != nil {
		return "", fmt.Errorf("failed to build tree for shotgun: %w", err)
	}

	if err := jobCtx.Err(); err != nil { // check for cancellation before final string operations
		return "", err
	}

	// the final output is the tree, a newline, then all concatenated file contents.
	// if filecontents is empty, we still want the newline after the tree.
	// if filecontents is not empty, it already ends with a newline, so an extra one might not be desired
	// depending on how it's structured. given each <file> block ends with \n, this should be fine.
	return output.String() + "\n" + strings.TrimRight(fileContents.String(), "\n"), nil
}

// --- watchman implementation ---

type Watchman struct {
	app         *App
	rootDir     string
	fsWatcher   *fsnotify.Watcher
	watchedDirs map[string]bool // tracks directories explicitly added to fsnotify

	// lastknownstate map[string]filemeta // removed, fsnotify handles state
	mu         sync.Mutex // changed to mutex for simplicity with start/stop/refresh
	cancelFunc context.CancelFunc

	// store current patterns to be used by scandirectorystateinternal
	currentProjectGitignore *gitignore.GitIgnore
	currentCustomPatterns   *gitignore.GitIgnore
}

func NewWatchman(app *App) *Watchman {
	return &Watchman{
		app:         app,
		watchedDirs: make(map[string]bool),
	}
}

// startfilewatcher is called by javascript to start watching a directory.
func (a *App) StartFileWatcher(rootDirPath string) error {
	runtime.LogInfof(a.ctx, "startfilewatcher called for: %s", rootDirPath)
	if a.fileWatcher == nil {
		return fmt.Errorf("file watcher not initialized")
	}
	return a.fileWatcher.Start(rootDirPath)
}

// stopfilewatcher is called by javascript to stop the current watcher.
func (a *App) StopFileWatcher() error {
	runtime.LogInfo(a.ctx, "stopfilewatcher called")
	if a.fileWatcher == nil {
		return fmt.Errorf("file watcher not initialized")
	}
	a.fileWatcher.Stop()
	return nil
}

func (w *Watchman) Start(newRootDir string) error {
	w.Stop() // stop any existing watcher

	w.mu.Lock()
	w.rootDir = newRootDir
	if w.rootDir == "" {
		w.mu.Unlock()
		runtime.LogInfo(w.app.ctx, "watchman: root directory is empty, not starting.")
		return nil
	}
	w.mu.Unlock()

	// initialize patterns based on app's current state
	if w.app.useGitignore {
		w.currentProjectGitignore = w.app.projectGitignore
	} else {
		w.currentProjectGitignore = nil
	}
	if w.app.useCustomIgnore {
		w.currentCustomPatterns = w.app.currentCustomIgnorePatterns
	} else {
		w.currentCustomPatterns = nil
	}

	w.mu.Lock()
	// ensure settings are loaded if they haven't been (e.g. if called before startup completes, though unlikely)
	// however, loadsettings is called in startup, so this should generally be populated.
	ctx, cancel := context.WithCancel(w.app.ctx) // use app's context as parent
	w.cancelFunc = cancel
	w.mu.Unlock()

	var err error
	w.fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		runtime.LogErrorf(w.app.ctx, "watchman: error creating fsnotify watcher: %v", err)
		return fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}
	w.watchedDirs = make(map[string]bool) // initialize/clear

	runtime.LogInfof(w.app.ctx, "watchman: starting for directory %s", newRootDir)
	w.addPathsToWatcherRecursive(newRootDir) // add initial paths

	go w.run(ctx)
	return nil
}

func (w *Watchman) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cancelFunc != nil {
		runtime.LogInfo(w.app.ctx, "watchman: stopping...")
		w.cancelFunc()
		w.cancelFunc = nil // allow gc and prevent double-cancel
	}
	if w.fsWatcher != nil {
		err := w.fsWatcher.Close()
		if err != nil {
			runtime.LogWarningf(w.app.ctx, "watchman: error closing fsnotify watcher: %v", err)
		}
		w.fsWatcher = nil
	}
	w.rootDir = ""
	w.watchedDirs = make(map[string]bool) // clear watched directories
}

func (w *Watchman) run(ctx context.Context) {
	defer func() {
		if w.fsWatcher != nil {
			// this close is a safeguard; stop() should ideally be called.
			w.fsWatcher.Close()
		}
		runtime.LogInfo(w.app.ctx, "watchman: goroutine stopped.")
	}()

	w.mu.Lock()
	currentRootDir := w.rootDir
	w.mu.Unlock()
	runtime.LogInfof(w.app.ctx, "watchman: monitoring goroutine started for %s", currentRootDir)

	for {
		select {
		case <-ctx.Done():
			w.mu.Lock()
			shutdownRootDir := w.rootDir // re-fetch rootdir under lock as it might have changed
			w.mu.Unlock()
			runtime.LogInfof(w.app.ctx, "watchman: context cancelled, shutting down watcher for %s.", shutdownRootDir)
			return

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				runtime.LogInfo(w.app.ctx, "watchman: fsnotify events channel closed.")
				return
			}
			runtime.LogDebugf(w.app.ctx, "watchman: fsnotify event: %s", event)

			w.mu.Lock()
			currentRootDir = w.rootDir // update currentrootdir under lock
			// safely copy ignore patterns
			projIgn := w.currentProjectGitignore
			custIgn := w.currentCustomPatterns
			w.mu.Unlock()

			if currentRootDir == "" { // watcher might have been stopped
				continue
			}

			relEventPath, err := filepath.Rel(currentRootDir, event.Name)
			if err != nil {
				runtime.LogWarningf(w.app.ctx, "watchman: could not get relative path for event %s (root: %s): %v", event.Name, currentRootDir, err)
				continue
			}

			// check if the event path is ignored
			isIgnoredByGit := projIgn != nil && projIgn.MatchesPath(relEventPath)
			isIgnoredByCustom := custIgn != nil && custIgn.MatchesPath(relEventPath)

			if isIgnoredByGit || isIgnoredByCustom {
				runtime.LogDebugf(w.app.ctx, "watchman: ignoring event for %s as it's an ignored path.", event.Name)
				continue
			}

			// handle relevant events (excluding chmod)
			if event.Op&fsnotify.Chmod == 0 {
				runtime.LogInfof(w.app.ctx, "watchman: relevant change detected for %s in %s", event.Name, currentRootDir)
				w.app.notifyFileChange(currentRootDir)
			}

			// dynamic directory watching
			if event.Op&fsnotify.Create != 0 {
				info, statErr := os.Stat(event.Name)
				if statErr == nil && info.IsDir() {
					// check if this new directory itself is ignored before adding
					isNewDirIgnoredByGit := projIgn != nil && projIgn.MatchesPath(relEventPath)
					isNewDirIgnoredByCustom := custIgn != nil && custIgn.MatchesPath(relEventPath)
					if !isNewDirIgnoredByGit && !isNewDirIgnoredByCustom {
						runtime.LogDebugf(w.app.ctx, "watchman: new directory created %s, adding to watcher.", event.Name)
						w.addPathsToWatcherRecursive(event.Name) // this will add event.name and its children
					} else {
						runtime.LogDebugf(w.app.ctx, "watchman: new directory %s is ignored, not adding to watcher.", event.Name)
					}
				}
			}

			if event.Op&fsnotify.Remove != 0 || event.Op&fsnotify.Rename != 0 {
				w.mu.Lock()
				if w.watchedDirs[event.Name] {
					runtime.LogDebugf(w.app.ctx, "watchman: watched directory %s removed/renamed, removing from watcher.", event.Name)
					// fsnotify might remove it automatically, but explicit removal is safer for our tracking
					if w.fsWatcher != nil { // check fswatcher as it might be closed by stop()
						err := w.fsWatcher.Remove(event.Name)
						if err != nil {
							runtime.LogWarningf(w.app.ctx, "watchman: error removing path %s from fsnotify: %v", event.Name, err)
						}
					}
					delete(w.watchedDirs, event.Name)
				}
				w.mu.Unlock()
			}

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				runtime.LogInfo(w.app.ctx, "watchman: fsnotify errors channel closed.")
				return
			}
			runtime.LogErrorf(w.app.ctx, "watchman: fsnotify error: %v", err)
		}
	}
}

func (w *Watchman) addPathsToWatcherRecursive(baseDirToAdd string) {
	w.mu.Lock() // lock to access watcher and ignore patterns
	fsW := w.fsWatcher
	projIgn := w.currentProjectGitignore
	custIgn := w.currentCustomPatterns
	overallRoot := w.rootDir
	w.mu.Unlock()

	if fsW == nil || overallRoot == "" {
		runtime.LogWarningf(w.app.ctx, "watchman.addpathstowatcherrecursive: fswatcher is nil or rootdir is empty. skipping add for %s.", baseDirToAdd)
		return
	}

	filepath.WalkDir(baseDirToAdd, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			runtime.LogWarningf(w.app.ctx, "watchman scan error accessing %s: %v", path, walkErr)
			if d != nil && d.IsDir() && path != overallRoot { // changed scanrootdir to overallroot for clarity
				return filepath.SkipDir
			}
			return nil // try to continue
		}

		if !d.IsDir() {
			return nil
		}

		relPath, errRel := filepath.Rel(overallRoot, path)
		if errRel != nil {
			runtime.LogWarningf(w.app.ctx, "watchman.addpathstowatcherrecursive: could not get relative path for %s (root: %s): %v", path, overallRoot, errRel)
			return nil // continue with other paths
		}

		// skip .git directory at the top level of overallroot
		if d.IsDir() && d.Name() == ".git" {
			parentDir := filepath.Dir(path)
			if parentDir == overallRoot {
				runtime.LogDebugf(w.app.ctx, "watchman.addpathstowatcherrecursive: skipping .git directory: %s", path)
				return filepath.SkipDir
			}
		}

		isIgnoredByGit := projIgn != nil && projIgn.MatchesPath(relPath)
		isIgnoredByCustom := custIgn != nil && custIgn.MatchesPath(relPath)

		if isIgnoredByGit || isIgnoredByCustom {
			runtime.LogDebugf(w.app.ctx, "watchman.addpathstowatcherrecursive: skipping ignored directory: %s", path)
			return filepath.SkipDir
		}

		errAdd := fsW.Add(path)
		if errAdd != nil {
			runtime.LogWarningf(w.app.ctx, "watchman.addpathstowatcherrecursive: error adding path %s to fsnotify: %v", path, errAdd)
		} else {
			runtime.LogDebugf(w.app.ctx, "watchman.addpathstowatcherrecursive: added to watcher: %s", path)
			w.mu.Lock()
			w.watchedDirs[path] = true
			w.mu.Unlock()
		}
		return nil
	})
}

// notifyfilechange is an internal method for the app to emit a wails event.
func (a *App) notifyFileChange(rootDir string) {
	runtime.EventsEmit(a.ctx, "projectFilesChanged", rootDir)
}

// refreshignoresandrescan is called when ignore settings change in the app.
func (w *Watchman) RefreshIgnoresAndRescan() error {
	w.mu.Lock()
	if w.rootDir == "" {
		w.mu.Unlock()
		runtime.LogInfo(w.app.ctx, "watchman.refreshignoresandrescan: no rootdir, skipping.")
		return nil
	}
	runtime.LogInfo(w.app.ctx, "watchman.refreshignoresandrescan: refreshing ignore patterns and re-scanning.")

	// update patterns based on app's current state
	if w.app.useGitignore {
		w.currentProjectGitignore = w.app.projectGitignore
	} else {
		w.currentProjectGitignore = nil
	}
	if w.app.useCustomIgnore {
		w.currentCustomPatterns = w.app.currentCustomIgnorePatterns
	} else {
		w.currentCustomPatterns = nil
	}
	currentRootDir := w.rootDir
	defer w.mu.Unlock()

	// stop existing watcher (closes, clears watcheddirs)
	if w.cancelFunc != nil {
		w.cancelFunc()
	}
	if w.fsWatcher != nil {
		w.fsWatcher.Close()
	}
	w.watchedDirs = make(map[string]bool)

	// create new watcher
	var err error
	w.fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		runtime.LogErrorf(w.app.ctx, "watchman.refreshignoresandrescan: error creating new fsnotify watcher: %v", err)
		return fmt.Errorf("failed to create new fsnotify watcher: %w", err)
	}

	w.addPathsToWatcherRecursive(currentRootDir) // add paths with new rules
	w.app.notifyFileChange(currentRootDir)       // notify frontend to refresh its view

	return nil
}

// --- configuration management ---

func (a *App) compileCustomIgnorePatterns() error {
	if strings.TrimSpace(a.settings.CustomIgnoreRules) == "" {
		a.currentCustomIgnorePatterns = nil
		runtime.LogDebug(a.ctx, "custom ignore rules are empty, no patterns compiled.")
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(a.settings.CustomIgnoreRules, "\r\n", "\n"), "\n")
	var validLines []string
	for _, line := range lines {
		// compileignorelines should handle empty/comment lines appropriately based on .gitignore syntax
		validLines = append(validLines, line)
	}

	ign := gitignore.CompileIgnoreLines(validLines...)
	// поскольку compileignorelines в этой версии не возвращает ошибку,
	// проверка на err удалена.
	// если ign будет nil (например, если все строки были пустыми или комментариями,
	// и библиотека так обрабатывает), то это будет корректно обработано ниже.
	a.currentCustomIgnorePatterns = ign
	runtime.LogInfo(a.ctx, "successfully compiled custom ignore patterns.")
	return nil
}

func (a *App) loadSettings() {
	// default to embedded rules
	a.settings.CustomIgnoreRules = defaultCustomIgnoreRulesContent

	if a.configPath == "" {
		runtime.LogWarningf(a.ctx, "config path is empty, using default custom ignore rules (embedded).")
		if err := a.compileCustomIgnorePatterns(); err != nil {
			// error already logged in compilecustomignorepatterns
		}
		return
	}

	data, err := os.ReadFile(a.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			runtime.LogInfo(a.ctx, "settings file not found. using default custom ignore rules (embedded) and attempting to save them.")
			// save default settings to create the file. compilecustomignorepatterns will be called after this.
			if errSave := a.saveSettings(); errSave != nil { // savesettings will use a.settings.customignorerules which is currently default
				runtime.LogErrorf(a.ctx, "failed to save default settings: %v", errSave)
			}
		} else {
			runtime.LogErrorf(a.ctx, "error reading settings file %s: %v. using default custom ignore rules (embedded).", a.configPath, err)
		}
	} else {
		var loadedSettings AppSettings
		err = json.Unmarshal(data, &loadedSettings)
		if err != nil {
			runtime.LogErrorf(a.ctx, "error unmarshalling settings from %s: %v. using default custom ignore rules (embedded).", a.configPath, err)
			a.settings.CustomIgnoreRules = defaultCustomIgnoreRulesContent // reset to default on unmarshal error
		} else {
			runtime.LogInfo(a.ctx, "successfully loaded custom rules from config. they will be combined with the embedded defaults.")
			// always start with the fresh embedded defaults, then append any user rules that are not already in the default.
			// this ensures that updates to the embedded ignore.glob are always included.
			baseRules := defaultCustomIgnoreRulesContent
			userRules := loadedSettings.CustomIgnoreRules

			// a simple model: combine and let the gitignore logic handle duplicates. last pattern wins.
			a.settings.CustomIgnoreRules = baseRules + "\n\n#--- user rules ---\n" + userRules

			// handle custompromptrules separately as it's a replacement, not an addition.
			if strings.TrimSpace(loadedSettings.CustomPromptRules) != "" {
				a.settings.CustomPromptRules = loadedSettings.CustomPromptRules
			} else {
				a.settings.CustomPromptRules = defaultCustomPromptRulesContent
			}
		}
	}

	if errCompile := a.compileCustomIgnorePatterns(); errCompile != nil {
		// error already logged in compilecustomignorepatterns
	}
}

func (a *App) saveSettings() error {
	if a.configPath == "" {
		err := errors.New("config path is not set, cannot save settings")
		runtime.LogError(a.ctx, err.Error())
		return err
	}

	data, err := json.MarshalIndent(a.settings, "", "  ")
	if err != nil {
		runtime.LogErrorf(a.ctx, "error marshalling settings: %v", err)
		return err
	}

	configDir := filepath.Dir(a.configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		runtime.LogErrorf(a.ctx, "error creating config directory %s: %v", configDir, err)
		return err
	}

	err = os.WriteFile(a.configPath, data, 0644)
	if err != nil {
		runtime.LogErrorf(a.ctx, "error writing settings to %s: %v", a.configPath, err)
		return err
	}
	runtime.LogInfo(a.ctx, "settings saved successfully.")
	return nil
}

// getcustomignorerules returns the current custom ignore rules as a string.
func (a *App) GetCustomIgnoreRules() string {
	// ensure settings are loaded if they haven't been (e.g. if called before startup completes, though unlikely)
	// however, loadsettings is called in startup, so this should generally be populated.
	return a.settings.CustomIgnoreRules
}

// setcustomignorerules updates the custom ignore rules, saves them, and recompiles.
func (a *App) SetCustomIgnoreRules(rules string) error {
	a.settings.CustomIgnoreRules = rules
	// attempt to compile first. if compilation fails, we might not want to save invalid rules,
	// or save them and let the user know they are not effective.
	// for now, compile then save. if compile fails, the old patterns (or nil) remain active.
	compileErr := a.compileCustomIgnorePatterns()

	saveErr := a.saveSettings()
	if saveErr != nil {
		return fmt.Errorf("failed to save settings: %w (compile error: %v)", saveErr, compileErr)
	}
	if compileErr != nil {
		return fmt.Errorf("rules saved, but failed to compile custom ignore patterns: %w", compileErr)
	}

	if a.fileWatcher != nil && a.fileWatcher.rootDir != "" {
		return a.fileWatcher.RefreshIgnoresAndRescan()
	}
	return nil
}

// getcustompromptrules returns the current custom prompt rules as a string.
func (a *App) GetCustomPromptRules() string {
	if strings.TrimSpace(a.settings.CustomPromptRules) == "" {
		return defaultCustomPromptRulesContent
	}
	return a.settings.CustomPromptRules
}

// setcustompromptrules updates the custom prompt rules and saves them.
func (a *App) SetCustomPromptRules(rules string) error {
	a.settings.CustomPromptRules = rules
	err := a.saveSettings()
	if err != nil {
		return fmt.Errorf("failed to save custom prompt rules: %w", err)
	}
	runtime.LogInfo(a.ctx, "custom prompt rules saved successfully.")
	return nil
}

// setusegitignore updates the app's setting for using .gitignore and informs the watcher.
func (a *App) SetUseGitignore(enabled bool) error {
	a.useGitignore = enabled
	runtime.LogInfof(a.ctx, "app setting usegitignore changed to: %v", enabled)
	if a.fileWatcher != nil && a.fileWatcher.rootDir != "" {
		// assuming watcher is for the current project if active.
		return a.fileWatcher.RefreshIgnoresAndRescan()
	}
	return nil
}

// setusecustomignore updates the app's setting for using custom ignore rules and informs the watcher.
func (a *App) SetUseCustomIgnore(enabled bool) error {
	a.useCustomIgnore = enabled
	runtime.LogInfof(a.ctx, "app setting usecustomignore changed to: %v", enabled)
	if a.fileWatcher != nil && a.fileWatcher.rootDir != "" {
		// assuming watcher is for the current project if active.
		return a.fileWatcher.RefreshIgnoresAndRescan()
	}
	return nil
}

// CountGeminiTokens counts the tokens in the provided text using Google's Gemini API
func (a *App) CountGeminiTokens(text string) (int, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return 0, fmt.Errorf("google_api_key environment variable not set")
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return 0, fmt.Errorf("failed to create genai client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro-latest")

	resp, err := model.CountTokens(context.Background(), genai.Text(text))
	if err != nil {
		return 0, fmt.Errorf("token counting failed: %w", err)
	}

	return int(resp.TotalTokens), nil
}

// ExecuteGeminiRequest sends a prompt to Google Gemini API
func (a *App) ExecuteGeminiRequest(prompt string, modelName string) (string, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return "", errors.New("google api key not set. please set the GOOGLE_API_KEY environment variable")
	}

	// create a context with cancellation capability
	ctx, cancel := context.WithCancel(a.ctx)
	defer cancel()

	// store the cancel function so it can be used by StopGeminiRequest
	a.geminiRequestCancel = cancel

	// create gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %v", err)
	}
	defer client.Close()

	// default to flash if model name is empty or unsupported
	if strings.TrimSpace(modelName) == "" {
		modelName = "gemini-2.5-flash"
	}

	model := client.GenerativeModel(modelName)
	model.SetTemperature(0.1)  // Set low temperature as per requirements

	// log request configuration and a preview of the request body (max 500 chars) for easier debugging
	bodyPreview := prompt
	if len(bodyPreview) > 500 {
		bodyPreview = bodyPreview[:500] + "..."
	}
	charCount := utf8.RuneCountInString(prompt)
	runtime.LogInfof(a.ctx, "gemini request config: model=%s temperature=0.1", modelName)
	runtime.LogInfof(a.ctx, "gemini request body length: %d characters", charCount)
	runtime.LogInfof(a.ctx, "gemini request body preview: %s", bodyPreview)

	// create chat session
	cs := model.StartChat()

	runtime.EventsEmit(a.ctx, "gemini_request_start", nil)

	// send the prompt and get response
	resp, err := cs.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		if errors.Is(err, context.Canceled) {
			runtime.EventsEmit(a.ctx, "gemini_request_canceled", nil)
			return "", errors.New("request was canceled")
		}
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	// extract text from response manually for compatibility with sdk versions lacking resp.Text()
	var sb strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					sb.WriteString(string(txt))
				}
			}
		}
	}
	respText := strings.TrimSpace(sb.String())

	// remove ```diff and closing ``` code block markers, replacing them with blank lines
	respText = strings.ReplaceAll(respText, "```diff", "\n")
	respText = strings.ReplaceAll(respText, "```", "\n")

	// trim any leading/trailing whitespace after cleanup
	respText = strings.TrimSpace(respText)

	if respText == "" {
		return "", errors.New("no response from gemini")
	}

	runtime.EventsEmit(a.ctx, "gemini_request_complete", nil)

	return respText, nil
}

// StopGeminiRequest cancels any ongoing Gemini API request
func (a *App) StopGeminiRequest() error {
	if a.geminiRequestCancel != nil {
		a.geminiRequestCancel()
		a.geminiRequestCancel = nil
		runtime.EventsEmit(a.ctx, "gemini_request_canceled", nil)
		return nil
	}
	return errors.New("no active gemini request to cancel")
}

// istextcontent heuristically determines whether the provided byte slice represents textual data.
// it checks for the presence of null bytes, utf-8 validity, and a ratio of non-printable control
// characters. this helps us avoid inlining binary data (e.g. images, audio) into the generated
// project context output. for non-text files we will emit a placeholder instead of raw bytes.
func isTextContent(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	// quick checks: null byte or invalid utf-8 => binary
	if bytes.IndexByte(data, 0) != -1 || !utf8.Valid(data) {
		return false
	}

	// sample first kib to estimate control char ratio
	sampleSize := len(data)
	if sampleSize > 1024 {
		sampleSize = 1024
	}

	nonPrintable := 0
	for i := 0; i < sampleSize; i++ {
		b := data[i]
		if b < 32 && b != 9 && b != 10 && b != 13 { // allow tab, lf, cr
			nonPrintable++
		}
	}
	// if more than 5% of sampled bytes are control characters, treat as binary
	return nonPrintable*20 <= sampleSize
}
