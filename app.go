package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore" // Import the gitignore library
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type FileNode struct {
	Name            string      `json:"name"`
	Path            string      `json:"path"`    // Full path
	RelPath         string      `json:"relPath"` // Path relative to selected root
	IsDir           bool        `json:"isDir"`
	Children        []*FileNode `json:"children,omitempty"`
	IsGitignored    bool        `json:"isGitignored"`    // True if path matches a .gitignore rule
	IsCustomIgnored bool        `json:"isCustomIgnored"` // True if path matches a ignore.glob rule
}

// SelectDirectory opens a dialog to select a directory and returns the chosen path
func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
}

// ListFiles lists files and folders in a directory, parsing .gitignore if present
func (a *App) ListFiles(dirPath string) ([]*FileNode, error) {
	var gitIgn *gitignore.GitIgnore
	gitignorePath := filepath.Join(dirPath, ".gitignore")
	fmt.Printf("Attempting to find .gitignore at: %s\n", gitignorePath)
	if _, err := os.Stat(gitignorePath); err == nil {
		fmt.Printf(".gitignore found at: %s\n", gitignorePath)
		gitIgn, err = gitignore.CompileIgnoreFile(gitignorePath)
		if err != nil {
			fmt.Printf("Error compiling .gitignore file at %s: %v\n", gitignorePath, err)
			gitIgn = nil // Ensure ign is nil if compilation fails
		} else {
			fmt.Printf(".gitignore compiled successfully.\n")
		}
	} else {
		fmt.Printf(".gitignore not found at %s (os.Stat error: %v)\n", gitignorePath, err)
		gitIgn = nil
	}

	var globIgn *gitignore.GitIgnore
	globIgnorePath := filepath.Join(dirPath, "ignore.glob")
	fmt.Printf("Attempting to find ignore.glob at: %s\n", globIgnorePath)
	if _, err := os.Stat(globIgnorePath); err == nil {
		fmt.Printf("ignore.glob found at: %s\n", globIgnorePath)
		globIgn, err = gitignore.CompileIgnoreFile(globIgnorePath)
		if err != nil {
			fmt.Printf("Error compiling ignore.glob file at %s: %v\n", globIgnorePath, err)
			globIgn = nil
		} else {
			fmt.Printf("ignore.glob compiled successfully.\n")
		}
	} else {
		fmt.Printf("ignore.glob not found at %s (os.Stat error: %v)\n", globIgnorePath, err)
		globIgn = nil
	}

	// Create the root node representing the selected directory
	rootNode := &FileNode{
		Name:    filepath.Base(dirPath),
		Path:    dirPath,
		RelPath: ".", // Relative path for the root itself is "."
		IsDir:   true,
		// IsGitignored for the root itself is typically false, unless a specific rule targets it.
		// For simplicity, we'll assume it's not ignored. If needed, this could be checked.
		IsGitignored: false,
	}

	// Get children for the root node using the existing buildTree logic
	children, err := buildTree(dirPath, dirPath, gitIgn, globIgn, 0)
	if err != nil {
		// If there's an error building the children tree (e.g., permission issues),
		// return the root node with no children, but also return the error.
		// Or, decide if this scenario means ListFiles should fail entirely.
		// For now, let's return the root and the error. The frontend might need to handle this.
		return []*FileNode{rootNode}, fmt.Errorf("error building children tree for %s: %w", dirPath, err)
	}
	rootNode.Children = children

	// ListFiles now returns a slice containing only the root node
	return []*FileNode{rootNode}, nil
}

func buildTree(currentPath, rootPath string, gitIgn *gitignore.GitIgnore, globIgn *gitignore.GitIgnore, depth int) ([]*FileNode, error) { // Added depth and globIgn
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}

	var nodes []*FileNode
	for _, entry := range entries {
		nodePath := filepath.Join(currentPath, entry.Name())
		relPath, _ := filepath.Rel(rootPath, nodePath)
		// For gitignore matching, paths should generally be relative to the .gitignore file (rootPath)
		// and use OS-specific separators. go-gitignore handles this.

		isGitignored := false
		isCustomIgnored := false
		pathToMatch := relPath
		if entry.IsDir() {
			if !strings.HasSuffix(pathToMatch, string(os.PathSeparator)) {
				pathToMatch += string(os.PathSeparator)
			}
		}

		if gitIgn != nil && gitIgn.MatchesPath(pathToMatch) {
			isGitignored = true
		}
		if globIgn != nil && globIgn.MatchesPath(pathToMatch) {
			isCustomIgnored = true
		}

		if depth < 2 || strings.Contains(relPath, "node_modules") || strings.HasSuffix(relPath, ".log") {
			fmt.Printf("Checking path: '%s' (original relPath: '%s'), IsDir: %v, Gitignored: %v, CustomIgnored: %v\n", pathToMatch, relPath, entry.IsDir(), isGitignored, isCustomIgnored)
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
			// Children inherit gitignore rules through their own path matching
			children, err := buildTree(nodePath, rootPath, gitIgn, globIgn, depth+1) // Increment depth, pass globIgn
			if err != nil {
				fmt.Printf("Error reading dir %s: %v\n", nodePath, err)
				continue
			}
			node.Children = children
		}
		nodes = append(nodes, node)
	}
	// Sort nodes: directories first, then files, then alphabetically
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

// GenerateShotgunOutput generates the TXT output
func (a *App) GenerateShotgunOutput(rootDir string, excludedPaths []string) (string, error) {
	var output strings.Builder
	var fileContents strings.Builder

	excludedMap := make(map[string]bool)
	for _, p := range excludedPaths {
		excludedMap[p] = true
	}

	var buildShotgunTree func(string, string, string) error
	buildShotgunTree = func(currentPath, prefix, rootDirRel string) error {
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return err
		}

		// Sort entries like in ListFiles for consistent tree
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

		// Create a temporary slice to hold non-excluded entries for correct prefixing
		var visibleEntries []fs.DirEntry
		for _, entry := range entries {
			path := filepath.Join(currentPath, entry.Name())
			relPath, _ := filepath.Rel(rootDir, path)
			if !excludedMap[relPath] {
				visibleEntries = append(visibleEntries, entry)
			}
		}

		for i, entry := range visibleEntries {
			path := filepath.Join(currentPath, entry.Name())
			// relPath is already computed above and checked against excludedMap
			relPath, _ := filepath.Rel(rootDir, path)

			isLast := i == len(visibleEntries)-1

			branch := "├── "
			nextPrefix := prefix + "│   "
			if isLast {
				branch = "└── "
				nextPrefix = prefix + "    "
			}
			output.WriteString(prefix + branch + entry.Name() + "\n")

			if entry.IsDir() {
				err := buildShotgunTree(path, nextPrefix, rootDirRel)
				if err != nil {
					fmt.Printf("Error processing subdirectory %s: %v\n", path, err)
				}
			} else {
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", path, err)
					content = []byte(fmt.Sprintf("Error reading file: %v", err))
				}

				var dirPart, filePart string
				// relPath is relative to rootDir (e.g., "src/main.go" or "file.txt")
				if strings.Contains(relPath, string(os.PathSeparator)) {
					dirPart = filepath.Dir(relPath) + string(os.PathSeparator)
					filePart = filepath.Base(relPath)
				} else {
					dirPart = "" // Root level file
					filePart = relPath
				}

				fileContents.WriteString(fmt.Sprintf("*#*#*%s%s*#*#*begin*#*#*\n", dirPart, filePart))
				fileContents.WriteString(string(content))
				fileContents.WriteString("\n*#*#*end*#*#*\n\n")
			}
		}
		return nil
	}

	output.WriteString(filepath.Base(rootDir) + string(os.PathSeparator) + "\n")

	err := buildShotgunTree(rootDir, "", rootDir)
	if err != nil {
		return "", fmt.Errorf("failed to build tree for shotgun: %w", err)
	}

	output.WriteString("\n")
	output.WriteString(fileContents.String())

	return output.String(), nil
}
