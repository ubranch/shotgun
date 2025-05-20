package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- Shotgun Diff Splitting ---

// ShotgunHunk represents a hunk within a shotgun diff file.
type ShotgunHunk struct {
	Range string `xml:"range,attr"`
	CDATA string `xml:",cdata"`
}

// ShotgunFileEntry represents a file entry within a shotgun diff.
type ShotgunFileEntry struct {
	Path  string        `xml:"path,attr"`
	Hunks []ShotgunHunk `xml:"hunk"`
}

// ShotgunDiff represents the structure of a shotgunDiff XML.
type ShotgunDiff struct {
	XMLName xml.Name           `xml:"shotgunDiff"`
	Xmlns   string             `xml:"xmlns,attr,omitempty"`
	Files   []ShotgunFileEntry `xml:"file"`
}

// SplitShotgunDiff parses a shotgunDiff XML string and splits it into multiple
// smaller shotgunDiff XML strings, each not exceeding approxLineLimit lines.
func (a *App) SplitShotgunDiff(diffXML string, approxLineLimit int) ([]string, error) {
	runtime.LogInfof(a.ctx, "SplitShotgunDiff called with line limit: %d", approxLineLimit)
	var parsedDiff ShotgunDiff
	if err := xml.Unmarshal([]byte(diffXML), &parsedDiff); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shotgunDiff XML: %w", err)
	}

	outputXmlns := parsedDiff.Xmlns
	if outputXmlns == "" {
		outputXmlns = "https://example.com/shotgun/diff/v1" // Default if not present
	}

	var splitDiffs []string
	currentLineCount := 0
	currentSplitFiles := make(map[string]*ShotgunFileEntry) // Map to store files in current split by path

	// Helper function to finalize the current split
	finalizeSplit := func() error {
		if len(currentSplitFiles) == 0 {
			return nil
		}

		processedPaths := make(map[string]bool) // Keep track of paths added
		currentSplitFileSlice := make([]ShotgunFileEntry, 0, len(currentSplitFiles))

		// First pass: iterate in the order of originalFile.Path to maintain original file order if possible
		for _, originalFile := range parsedDiff.Files {
			if entry, exists := currentSplitFiles[originalFile.Path]; exists {
				hunksCopy := make([]ShotgunHunk, len(entry.Hunks))
				copy(hunksCopy, entry.Hunks)
				currentSplitFileSlice = append(currentSplitFileSlice, ShotgunFileEntry{
					Path:  entry.Path,
					Hunks: hunksCopy,
				})
				processedPaths[originalFile.Path] = true
			}
		}

		// Second pass: add any files that might be in currentSplitFiles but were not processed
		// This ensures all items in currentSplitFiles are included, even if a path was not in parsedDiff.Files (unlikely)
		// or if just to iterate remaining items.
		if len(processedPaths) < len(currentSplitFiles) {
			runtime.LogDebugf(a.ctx, "finalizeSplit: %d files in current split, %d processed by original order. Adding remaining.", len(currentSplitFiles), len(processedPaths))
			// Iterate over the map keys for any remaining files. Order for these is map iteration order.
			for path, entry := range currentSplitFiles {
				if !processedPaths[path] {
					hunksCopy := make([]ShotgunHunk, len(entry.Hunks))
					copy(hunksCopy, entry.Hunks)
					currentSplitFileSlice = append(currentSplitFileSlice, ShotgunFileEntry{
						Path:  path, // Use path from map key
						Hunks: hunksCopy,
					})
				}
			}
		}

		currentOutput := ShotgunDiff{Xmlns: outputXmlns, Files: currentSplitFileSlice}
		xmlBytes, err := xml.MarshalIndent(currentOutput, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal split diff: %w", err)
		}
		splitDiffs = append(splitDiffs, string(xmlBytes))

		// Reset for next split
		currentSplitFiles = make(map[string]*ShotgunFileEntry)
		currentLineCount = 0
		return nil
	}

	for _, fileEntry := range parsedDiff.Files { // Iterate through each file in the input diff
		filePath := fileEntry.Path

		for _, hunk := range fileEntry.Hunks { // Iterate through each hunk of the current file
			hunkContentLines := strings.Split(strings.ReplaceAll(hunk.CDATA, "\r\n", "\n"), "\n")
			numLinesInHunk := countLinesInHunk(a, hunk, hunkContentLines, filePath)

			// Case 1: This hunk ALONE is larger than the approxLineLimit.
			// It must go into its own split.
			// If the current split has any content, finalize it first.
			if numLinesInHunk > approxLineLimit {
				if currentLineCount > 0 {
					if err := finalizeSplit(); err != nil {
						return nil, err
					}
				}
				// This large hunk forms a new split by itself.
				currentSplitFiles[filePath] = &ShotgunFileEntry{
					Path:  filePath,
					Hunks: []ShotgunHunk{hunk},
				}
				currentLineCount = numLinesInHunk // This will be > approxLineLimit

				if err := finalizeSplit(); err != nil { // Finalize this single-hunk split
					return nil, err
				}
				continue // Move to the next hunk in the input file
			}

			// Case 2: Adding this hunk would exceed the limit (but hunk itself is not > limit).
			// Finalize the current split (which does NOT include this hunk yet).
			// Then, this hunk will be added to a new, empty split.
			if currentLineCount > 0 && (currentLineCount+numLinesInHunk > approxLineLimit) {
				if err := finalizeSplit(); err != nil {
					return nil, err
				}
				// currentLineCount is now 0, currentSplitFiles is empty.
			}

			// Case 3: Add this hunk to the current split.
			// This occurs if:
			// a) The split was just reset (currentLineCount == 0).
			// b) Or, currentLineCount > 0 and adding this hunk does NOT exceed the limit.
			fileEntryInCurrentSplit, exists := currentSplitFiles[filePath]
			if !exists {
				fileEntryInCurrentSplit = &ShotgunFileEntry{
					Path:  filePath,
					Hunks: []ShotgunHunk{},
				}
				currentSplitFiles[filePath] = fileEntryInCurrentSplit
			}

			fileEntryInCurrentSplit.Hunks = append(fileEntryInCurrentSplit.Hunks, hunk)
			currentLineCount += numLinesInHunk
		}
	}

	// Add any remaining content as the final split
	if err := finalizeSplit(); err != nil {
		return nil, err
	}

	runtime.LogInfof(a.ctx, "Successfully split diff into %d parts.", len(splitDiffs))
	return splitDiffs, nil
}

// Helper function to count lines in a hunk.
// For splitting purposes, we count all lines in the CDATA block.
func countLinesInHunk(a *App, hunk ShotgunHunk, hunkContentLines []string, filePath string) int {
	numLinesInHunk := len(hunkContentLines)

	// Ensure that a hunk, especially if its CDATA is effectively empty but it's still processed as a hunk,
	// contributes at least 1 to the line count.
	// len(hunkContentLines) will be 1 for an empty CDATA string ("") due to strings.Split behavior.
	// This check is a safeguard for potential edge cases or if hunkContentLines could somehow be empty.
	if numLinesInHunk < 1 {
		numLinesInHunk = 1
		runtime.LogDebugf(a.ctx, "Hunk for file %s (range: %s) had calculated CDATA line count < 1, defaulting to 1.", filePath, hunk.Range)
	}

	runtime.LogDebugf(a.ctx, "Counted %d lines for hunk in file %s (range: %s) based on CDATA length (len(hunkContentLines)).", numLinesInHunk, filePath, hunk.Range)
	return numLinesInHunk
}

// StartupTest initializes the app for testing
// Note: This function remains here for now as it initializes the App struct for testing,
// which includes components beyond just diff splitting. It might be moved later
// if a more dedicated testing setup is introduced.
func (a *App) StartupTest(ctx context.Context) {
	a.ctx = ctx
	a.contextGenerator = NewContextGenerator(a)
	a.fileWatcher = NewWatchman(a)
	a.settings.CustomIgnoreRules = defaultCustomIgnoreRulesContent
	a.settings.CustomPromptRules = defaultCustomPromptRulesContent
	_ = a.compileCustomIgnorePatterns()
} 