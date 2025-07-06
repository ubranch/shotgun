package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// --- shotgun diff splitting ---

// splitshotgundiff parses a git diff string and splits it into multiple
// smaller git diff strings, each not exceeding approxlinelimit lines.
// it tries to split between file diffs first, then between chunks if a single file diff is too large.
func (a *App) SplitShotgunDiff(gitDiffText string, approxLineLimit int) ([]string, error) {
	runtime.LogInfof(a.ctx, "splitshotgundiff called with line limit: %d for git diff text", approxLineLimit)

	if strings.TrimSpace(gitDiffText) == "" {
		return []string{}, nil
	}

	// regex to find the start of a file diff block.
	// go's regex engine (re2) does not support lookarounds like (?=).
	// we will find the start indices of each file block and split manually.
	fileDiffStartRegex := regexp.MustCompile(`(?m)^diff --git `)
	startIndices := fileDiffStartRegex.FindAllStringIndex(gitDiffText, -1)

	var fileDiffBlocks []string

	if len(startIndices) == 0 {
		// if no "diff --git" is found, treat the whole input as a single block
		runtime.LogWarning(a.ctx, fmt.Sprintf("splitshotgundiff: no 'diff --git' blocks found in input. treating as single block."))
		if strings.TrimSpace(gitDiffText) != "" {
			fileDiffBlocks = append(fileDiffBlocks, gitDiffText)
		}
	} else {
		// split the text based on the start indices
		for i := 0; i < len(startIndices); i++ {
			start := startIndices[i][0]
			end := len(gitDiffText)
			if i+1 < len(startIndices) {
				end = startIndices[i+1][0]
			}
			block := gitDiffText[start:end]
			block = strings.TrimSpace(block)
			if block != "" {
				fileDiffBlocks = append(fileDiffBlocks, block)
			}
		}
	}

	var splitDiffs []string
	var currentSplitContent strings.Builder
	currentSplitLines := 0

	hunkHeaderRegex := regexp.MustCompile(`^@@ .* @@`)

	for _, fileBlock := range fileDiffBlocks {
		// fileblock is already trimmed by the splitting logic above, but continue check is fine
		if fileBlock == "" { continue }

		fileBlockLines := strings.Split(fileBlock, "\n")
		numLinesInFileBlock := len(fileBlockLines)

		// check if the fileblock itself is too large
		if numLinesInFileBlock > approxLineLimit {
			// if there's pending content in currentsplitcontent, finalize it
			if currentSplitContent.Len() > 0 {
				splitDiffs = append(splitDiffs, currentSplitContent.String())
				currentSplitContent.Reset()
				currentSplitLines = 0
			}

			// this fileblock is too large, needs to be split by chunks
			// extract file header (lines before the first hunk)
			firstHunkIndex := -1
			for i, line := range fileBlockLines {
				if hunkHeaderRegex.MatchString(line) {
					firstHunkIndex = i
					break
				}
			}

			if firstHunkIndex == -1 { // no chunks found, but block is large? unusual. treat as one large piece.
				runtime.LogWarning(a.ctx, fmt.Sprintf("splitshotgundiff: large file block without chunks in '%s'. treating as single block.", getPathFromDiffHeader(fileBlockLines[0])))
				splitDiffs = append(splitDiffs, fileBlock+"\n") // add newline for consistency if it's a full block
				continue
			}

			fileHeader := strings.Join(fileBlockLines[:firstHunkIndex], "\n") + "\n"
			numLinesInHeader := firstHunkIndex

			var currentFileSplitHunks strings.Builder
			currentFileSplitHunkLines := 0

			hunkStartIndex := firstHunkIndex
			for hunkStartIndex < len(fileBlockLines) {
				// find the end of the current hunk
				hunkEndIndex := hunkStartIndex + 1
				for hunkEndIndex < len(fileBlockLines) && !hunkHeaderRegex.MatchString(fileBlockLines[hunkEndIndex]) {
					hunkEndIndex++
				}

				currentHunkContent := strings.Join(fileBlockLines[hunkStartIndex:hunkEndIndex], "\n")
				numLinesInCurrentHunk := hunkEndIndex - hunkStartIndex

				// if this single hunk (plus header) is larger than limit, it gets its own split
				if numLinesInHeader + numLinesInCurrentHunk > approxLineLimit && currentFileSplitHunkLines == 0 {
					splitDiffs = append(splitDiffs, fileHeader + currentHunkContent + "\n")
					hunkStartIndex = hunkEndIndex
					continue
				}

				// if adding this hunk exceeds the limit (for this file's partial split)
				if currentFileSplitHunkLines > 0 && (numLinesInHeader + currentFileSplitHunkLines + numLinesInCurrentHunk > approxLineLimit) {
					splitDiffs = append(splitDiffs, fileHeader + currentFileSplitHunks.String())
					currentFileSplitHunks.Reset()
					currentFileSplitHunkLines = 0
				}

				currentFileSplitHunks.WriteString(currentHunkContent + "\n")
				currentFileSplitHunkLines += numLinesInCurrentHunk
				hunkStartIndex = hunkEndIndex
			}

			// add any remaining chunks for the current file
			if currentFileSplitHunks.Len() > 0 {
				splitDiffs = append(splitDiffs, fileHeader + currentFileSplitHunks.String())
			}

		} else { // file block is not too large by itself
			// if adding this fileblock would exceed the limit for the current_split
			if currentSplitLines > 0 && (currentSplitLines + numLinesInFileBlock > approxLineLimit) {
				splitDiffs = append(splitDiffs, currentSplitContent.String())
				currentSplitContent.Reset()
				currentSplitLines = 0
			}
			currentSplitContent.WriteString(fileBlock + "\n") // add newline between file blocks
			currentSplitLines += numLinesInFileBlock
		}
	}

	// add any remaining content as the final split
	if currentSplitContent.Len() > 0 {
		splitDiffs = append(splitDiffs, currentSplitContent.String())
	}

	// trim trailing newlines from each split diff for consistency and prepare for potential merging
	initialSplitDiffs := make([]string, 0, len(splitDiffs))
	initialSplitSizes := make([]int, 0, len(splitDiffs))
	for _, sDiff := range splitDiffs {
		trimmedDiff := strings.TrimSpace(sDiff)
		if trimmedDiff != "" {
			initialSplitDiffs = append(initialSplitDiffs, trimmedDiff)
			initialSplitSizes = append(initialSplitSizes, len(strings.Split(trimmedDiff, "\n")))
		}
	}

	// --- advanced merging logic ---
	// if approxlinelimit is not positive, merging logic is skipped.
	if approxLineLimit <= 0 {
		runtime.LogInfof(a.ctx, "approxlinelimit is %d, skipping merge step. returning %d initial splits.", approxLineLimit, len(initialSplitDiffs))
		return initialSplitDiffs, nil
	}

	// if there's 0 or 1 split, no merging is possible or needed.
	if len(initialSplitDiffs) <= 1 {
		runtime.LogInfof(a.ctx, "only %d initial split(s), no merging needed. returning as is.", len(initialSplitDiffs))
		return initialSplitDiffs, nil
	}

	runtime.LogInfof(a.ctx, "starting advanced merge step for %d initial splits with approxlinelimit %d.", len(initialSplitDiffs), approxLineLimit)

	// allow merged splits to be up to 20% larger than the user's approximate line limit.
	maxAllowedLines := int(float64(approxLineLimit) * 1.20)
	runtime.LogInfof(a.ctx, "max allowed lines per merged split: %d", maxAllowedLines)

	// this is a modified bin packing problem approach:
	// 1. initialize splitstomerge list with initial splits
	// 2. define a cost function to evaluate merged solutions
	// 3. try various combinations, picking the best solution

	type MergeGroup struct {
		Splits    []string
		LineCount int
	}

	// first, identify large splits that must be their own group as they're already close to or exceeding the limit
	var largeSplits []MergeGroup
	var smallSplits []int // indices of small splits we'll try to recombine

	for i, size := range initialSplitSizes {
		if size >= approxLineLimit { // already close to or above line limit - keep as is
			largeSplits = append(largeSplits, MergeGroup{
				Splits:    []string{initialSplitDiffs[i]},
				LineCount: size,
			})
			runtime.LogInfof(a.ctx, "split %d with %d lines kept as standalone group (already large)", i, size)
		} else {
			smallSplits = append(smallSplits, i)
		}
	}

	// if no small splits, return the identified large splits as-is
	if len(smallSplits) == 0 {
		runtime.LogInfof(a.ctx, "no small splits to merge, returning %d large splits as-is", len(largeSplits))
		result := make([]string, len(largeSplits))
		for i, group := range largeSplits {
			result[i] = group.Splits[0] // each large split is its own group with one split
		}
		return result, nil
	}

	// for small splits, try to find the optimal combination
	smallSplitData := make([]struct {
		Content   string
		LineCount int
	}, len(smallSplits))

	for i, idx := range smallSplits {
		smallSplitData[i].Content = initialSplitDiffs[idx]
		smallSplitData[i].LineCount = initialSplitSizes[idx]
	}

	// helper function to calculate solution score (lower is better)
	// prefers fewer groups and groups closer to maxallowedlines in size
	calculateSolutionScore := func(solution []MergeGroup) float64 {
		if len(solution) == 0 {
			return float64(1<<31 - 1) // maximum value, invalid solution
		}

		score := float64(len(solution)) * 1000 // base score is number of groups * 1000

		// add penalties for uneven groups and groups far below the limit
		for _, group := range solution {
			// penalty for how far the group is from the ideal size (maxallowedlines)
			// we prefer groups to be closer to maxallowedlines, but not over
			utilization := float64(group.LineCount) / float64(maxAllowedLines)
			if utilization > 1.0 {
				// severe penalty for exceeding max allowed lines
				score += 10000 * (utilization - 1.0)
			} else {
				// penalty for underutilization
				score += 100 * (1.0 - utilization)
			}
		}

		return score
	}

	// create initial solution with each small split in its own group
	initialSolution := make([]MergeGroup, len(smallSplitData))
	for i, data := range smallSplitData {
		initialSolution[i] = MergeGroup{
			Splits:    []string{data.Content},
			LineCount: data.LineCount,
		}
	}

	// apply a greedy bottom-up algorithm to merge small splits
	// try to select pairs of groups to merge, prioritizing those that give the best improvement in score
	currentSolution := initialSolution

	for {
		bestScore := calculateSolutionScore(currentSolution)
		var bestMerge struct {
			GroupIndex1 int
			GroupIndex2 int
			NewScore    float64
		}
		bestMerge.NewScore = bestScore
		mergeFound := false

		// try combining each pair of groups
		for i := 0; i < len(currentSolution); i++ {
			for j := i + 1; j < len(currentSolution); j++ {
				// check if merging is valid (doesn't exceed limits)
				// +1 for the newline separator between diffs
				combinedLineCount := currentSolution[i].LineCount + currentSolution[j].LineCount + 1
				if combinedLineCount <= maxAllowedLines {
					// try the merge and evaluate
					newSolution := make([]MergeGroup, 0, len(currentSolution) - 1)

					// add the merged group
					merged := MergeGroup{
						Splits:    append(append([]string{}, currentSolution[i].Splits...), currentSolution[j].Splits...),
						LineCount: combinedLineCount,
					}
					newSolution = append(newSolution, merged)

					// add all other groups
					for k := 0; k < len(currentSolution); k++ {
						if k != i && k != j {
							newSolution = append(newSolution, currentSolution[k])
						}
					}

					newScore := calculateSolutionScore(newSolution)
					if newScore < bestMerge.NewScore {
						bestMerge.GroupIndex1 = i
						bestMerge.GroupIndex2 = j
						bestMerge.NewScore = newScore
						mergeFound = true
					}
				}
			}
		}

		// if no improvement was found, stop
		if !mergeFound || bestMerge.NewScore >= bestScore {
			break
		}

		// apply the best merge
		i, j := bestMerge.GroupIndex1, bestMerge.GroupIndex2
		if i > j {
			i, j = j, i // ensure i < j to simplify logic below
		}

		// merge group j into group i
		combinedLineCount := currentSolution[i].LineCount + currentSolution[j].LineCount + 1
		currentSolution[i].Splits = append(currentSolution[i].Splits, currentSolution[j].Splits...)
		currentSolution[i].LineCount = combinedLineCount

		// remove group j
		currentSolution = append(currentSolution[:j], currentSolution[j+1:]...)

		runtime.LogInfof(a.ctx, "merged two groups, solution now has %d groups with score %.2f",
			len(currentSolution), bestMerge.NewScore)
	}

	// combine the large splits and the optimized small splits
	finalGroups := append(largeSplits, currentSolution...)
	runtime.LogInfof(a.ctx, "final solution: %d groups (%d large, %d optimized small groups)",
		len(finalGroups), len(largeSplits), len(currentSolution))

	// build the final result strings
	mergedSplitsResult := make([]string, len(finalGroups))
	for i, group := range finalGroups {
		if len(group.Splits) == 1 {
			// single split, no need to join
			mergedSplitsResult[i] = group.Splits[0]
		} else {
			// multiple splits, join with newlines
			mergedSplitsResult[i] = strings.Join(group.Splits, "\n")
		}
		runtime.LogInfof(a.ctx, "group %d: %d splits, %d lines", i, len(group.Splits), group.LineCount)
	}

	runtime.LogInfof(a.ctx, "split git diff: %d initial splits, merged into %d final splits. target line limit ~%d (merged max %d).",
		len(initialSplitDiffs), len(mergedSplitsResult), approxLineLimit, maxAllowedLines)
	return mergedSplitsResult, nil
}

// helper to get a/path from "diff --git a/path b/path"
func getPathFromDiffHeader(diffHeaderLine string) string {
	parts := strings.Fields(diffHeaderLine)
	if len(parts) >= 3 {
		return parts[2] // a/path
	}
	return "unknown_file"
}

// startuptest initializes the app for testing
// this is a minimal setup and should be expanded
func (a *App) StartupTest(ctx context.Context) {
	a.ctx = ctx
	a.contextGenerator = NewContextGenerator(a)
	a.fileWatcher = NewWatchman(a)
	a.settings.CustomIgnoreRules = defaultCustomIgnoreRulesContent
	a.settings.CustomPromptRules = defaultCustomPromptRulesContent
	_ = a.compileCustomIgnorePatterns()
}
