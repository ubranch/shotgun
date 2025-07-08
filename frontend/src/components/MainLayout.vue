<template>
    <div
        class="flex flex-col h-screen bg-light-bg dark:bg-[#141414] text-light-fg dark:text-dark-fg"
    >
        <div
            class="flex items-center justify-center bg-light-surface dark:bg-[#141414] border-b border-light-border dark:border-dark-border"
        >
            <HorizontalStepper
                :current-step="currentStep"
                :steps="steps"
                @navigate="navigateToStep"
                :key="`hstepper-${currentStep}-${steps
                    .map((s) => s.completed)
                    .join('')}`"
            />
        </div>
        <div class="flex flex-1 overflow-hidden">
            <LeftSidebar
                :current-step="currentStep"
                :steps="steps"
                :project-root="projectRoot"
                :file-tree-nodes="fileTree"
                :use-gitignore="useGitignore"
                :use-custom-ignore="useCustomIgnore"
                :loading-error="loadingError"
                @navigate="navigateToStep"
                @select-folder="selectProjectFolderHandler"
                @toggle-gitignore="toggleGitignoreHandler"
                @toggle-custom-ignore="toggleCustomIgnoreHandler"
                @toggle-exclude="toggleExcludeNode"
                @custom-rules-updated="handleCustomRulesUpdated"
                @select-all-files="selectAllFiles"
                @deselect-all-files="deselectAllFiles"
                @reset-file-selections="resetFileSelections"
                @add-log="({ message, type }) => addLog(message, type)"
            />
            <CentralPanel
                :current-step="currentStep"
                :shotgun-prompt-context="shotgunPromptContext"
                :generation-progress="generationProgressData"
                :is-generating-context="isGeneratingContext"
                :project-root="projectRoot"
                :platform="platform"
                :user-task="userTask"
                :rules-content="rulesContent"
                :split-diffs="splitDiffs"
                :is-loading-split-diffs="isLoadingSplitDiffs"
                :final-prompt="finalPrompt"
                :split-line-limit="splitLineLimitValue"
                :shotgun-git-diff="shotgunGitDiff"
                :split-line-limit-value="splitLineLimitValue"
                @step-action="handleStepAction"
                @update-composed-prompt="handleComposedPromptUpdate"
                @update:user-task="handleUserTaskUpdate"
                @update:rules-content="handleRulesContentUpdate"
                @update:shotgunGitDiff="handleShotgunGitDiffUpdate"
                @update:splitLineLimit="handleSplitLineLimitUpdate"
                ref="centralPanelRef"
            />
        </div>
    </div>
</template>

<script setup>
import {
    ref,
    reactive,
    watch,
    onMounted,
    onBeforeUnmount,
    nextTick,
} from "vue";
import HorizontalStepper from "./HorizontalStepper.vue";
import LeftSidebar from "./LeftSidebar.vue";
import CentralPanel from "./CentralPanel.vue";
import ThemeToggle from "./ThemeToggle.vue";
import {
    ListFiles,
    RequestShotgunContextGeneration,
    SelectDirectory as SelectDirectoryGo,
    StartFileWatcher,
    StopFileWatcher,
    SetUseGitignore,
    SetUseCustomIgnore,
    SplitShotgunDiff,
} from "../../wailsjs/go/main/App";
import { EventsOn, Environment } from "../../wailsjs/runtime/runtime";

// unlisten function for the auto-open-folder event emitted by the backend
let unlistenAutoOpenFolder = null;

const currentStep = ref(1);
const steps = ref([
    {
        id: 1,
        title: "prepare context",
        completed: false,
        description:
            "select project folder, review files, and generate the initial project context for the llm.",
    },
    {
        id: 2,
        title: "compose prompt",
        completed: false,
        description:
            "provide a prompt to the llm based on the project context to generate a code diff.",
    },
    {
        id: 3,
        title: "execute prompt",
        completed: false,
        description:
            "paste a large shotgundiff and split it into smaller, manageable parts.",
    },
    {
        id: 4,
        title: "apply patch",
        completed: false,
        description: "copy and apply the smaller diff parts to your project.",
    },
]);

const logMessages = ref([]);
const centralPanelRef = ref(null);

function addLog(message, type = "info", targetConsole = "bottom") {
    const logEntry = {
        message,
        type,
        timestamp: new Date().toLocaleTimeString(),
    };

    if (targetConsole === "bottom" || targetConsole === "both") {
        logMessages.value.push(logEntry);
    }
    if (targetConsole === "step" || targetConsole === "both") {
        if (
            centralPanelRef.value &&
            currentStep.value === 3 &&
            centralPanelRef.value.addLogToStep3Console
        ) {
            centralPanelRef.value.addLogToStep3Console(message, type);
        }
    }
}

const projectRoot = ref("");
const fileTree = ref([]);
const shotgunPromptContext = ref("");
const loadingError = ref("");
const useGitignore = ref(true);
const useCustomIgnore = ref(true);
const manuallyToggledNodes = reactive(new Map());
const isGeneratingContext = ref(false);
const generationProgressData = ref({ current: 0, total: 0 });
const isFileTreeLoading = ref(false);
const composedLlmPrompt = ref(""); // to store the prompt from step 2
const platform = ref("unknown"); // to store os platform (e.g., 'darwin', 'windows', 'linux')
const userTask = ref("");
const rulesContent = ref("");
const finalPrompt = ref("");
const isLoadingSplitDiffs = ref(false);
const splitDiffs = ref([]);
const shotgunGitDiff = ref("");
const splitLineLimitValue = ref(0); // add new state variable
let debounceTimer = null;

// watcher related
const projectFilesChangedPendingReload = ref(false);
let unlistenProjectFilesChanged = null;

async function selectProjectFolderHandler() {
    isFileTreeLoading.value = true;
    try {
        shotgunPromptContext.value = "";
        isGeneratingContext.value = false;
        const selectedDir = await SelectDirectoryGo();
        if (selectedDir) {
            projectRoot.value = selectedDir;
            loadingError.value = "";
            manuallyToggledNodes.clear();
            fileTree.value = [];

            await loadFileTree(selectedDir);

            splitDiffs.value = []; // clear any previous splits when new project selected

            if (!isFileTreeLoading.value && projectRoot.value) {
                debouncedTriggerShotgunContextGeneration();
            }

            steps.value.forEach((s) => (s.completed = false));
            currentStep.value = 1;
            addLog(`project folder selected: ${selectedDir}`, "info", "bottom");
        } else {
            isFileTreeLoading.value = false;
        }
    } catch (err) {
        console.error("error selecting directory:", err);
        const errorMsg = "failed to select directory: " + (err.message || err);
        loadingError.value = errorMsg;
        addLog(errorMsg, "error", "bottom");
        isFileTreeLoading.value = false;
    }
}

async function loadFileTree(dirPath) {
    isFileTreeLoading.value = true;
    loadingError.value = "";
    addLog(`loading file tree for: ${dirPath}`, "info", "bottom");
    try {
        const treeData = await ListFiles(dirPath);
        fileTree.value = mapDataToTreeRecursive(treeData, null);
        addLog(
            `file tree loaded successfully. root items: ${fileTree.value.length}`,
            "info",
            "bottom"
        );
    } catch (err) {
        console.error("error listing files:", err);
        const errorMsg = "failed to load file tree: " + (err.message || err);
        loadingError.value = errorMsg;
        addLog(errorMsg, "error", "bottom");
        fileTree.value = [];
    } finally {
        isFileTreeLoading.value = false;
        checkAndProcessPendingFileTreeReload();
    }
}

function calculateNodeExcludedState(node) {
    const manualToggle = manuallyToggledNodes.get(node.relPath);
    if (manualToggle !== undefined) return manualToggle;
    if (useGitignore.value && node.isGitignored) return true;
    if (useCustomIgnore.value && node.isCustomIgnored) return true;
    return false;
}

function mapDataToTreeRecursive(nodes, parent) {
    if (!nodes) return [];
    return nodes.map((node) => {
        const isRootNode = parent === null;
        const reactiveNode = reactive({
            ...node,
            expanded: node.isDir ? isRootNode : undefined,
            parent: parent,
            children: [],
        });
        reactiveNode.excluded = calculateNodeExcludedState(reactiveNode);

        if (node.children && node.children.length > 0) {
            reactiveNode.children = mapDataToTreeRecursive(
                node.children,
                reactiveNode
            );
        }
        return reactiveNode;
    });
}

function isAnyParentVisuallyExcluded(node) {
    if (!node || !node.parent) {
        return false;
    }
    let current = node.parent;
    while (current) {
        if (current.excluded) {
            // current.excluded reflects its visual/checkbox state
            return true;
        }
        current = current.parent;
    }
    return false;
}

function toggleExcludeNode(nodeToToggle) {
    // if the node is under an unselected parent and is currently unselected itself (nodetotoggle.excluded is true),
    // the first click should select it (set nodetotoggle.excluded to false).
    if (isAnyParentVisuallyExcluded(nodeToToggle) && nodeToToggle.excluded) {
        nodeToToggle.excluded = false;
    } else {
        // otherwise, normal toggle behavior.
        nodeToToggle.excluded = !nodeToToggle.excluded;
    }
    manuallyToggledNodes.set(nodeToToggle.relPath, nodeToToggle.excluded);

    // propagate this manual toggle to all descendants so that child checkboxes
    // follow the parent folder state. this fixes the scenario where "deselect all"
    // followed by selecting a folder would leave its children unchecked.
    function propagateToDescendants(node, newExcluded) {
        if (!node.children || node.children.length === 0) return;
        node.children.forEach(child => {
            if (newExcluded) {
                // when excluding, force child exclusion via manual toggle
                child.excluded = true;
                manuallyToggledNodes.set(child.relPath, true);
            } else {
                // when including, remove any manual override so ignore rules can apply
                manuallyToggledNodes.delete(child.relPath);
            }
            propagateToDescendants(child, newExcluded);
        });
    }

    propagateToDescendants(nodeToToggle, nodeToToggle.excluded);

    // immediately propagate the visual exclusion state through the subtree so that
    // child checkboxes reflect the folder selection change without waiting for
    // the deep watcher tick.
    updateAllNodesExcludedState(fileTree.value);

    addLog(
        `toggled exclusion for ${nodeToToggle.name} to ${nodeToToggle.excluded} (children updated)`,
        "info",
        "bottom"
    );
}

function updateAllNodesExcludedState(nodesToUpdate) {
    // this is the public-facing function
    // it calls the recursive helper, starting with parentisvisuallyexcluded = false for root nodes.
    _updateAllNodesExcludedStateRecursive(nodesToUpdate, false);
}

function _updateAllNodesExcludedStateRecursive(
    nodesToUpdate,
    parentIsVisuallyExcluded
) {
    if (!nodesToUpdate || nodesToUpdate.length === 0) return;
    nodesToUpdate.forEach((node) => {
        const manualToggle = manuallyToggledNodes.get(node.relPath);
        let isExcludedByRule = false;
        if (useGitignore.value && node.isGitignored) isExcludedByRule = true;
        if (useCustomIgnore.value && node.isCustomIgnored)
            isExcludedByRule = true;

        if (manualToggle !== undefined) {
            // if there's a manual toggle, it dictates the state.
            node.excluded = manualToggle;
        } else {
            // if not manually toggled, it's excluded if a rule matches or if its parent is visually excluded.
            // this establishes the default inherited exclusion for visual purposes.
            node.excluded = isExcludedByRule || parentIsVisuallyExcluded;
        }

        if (node.children && node.children.length > 0) {
            _updateAllNodesExcludedStateRecursive(node.children, node.excluded); // pass current node's new visual excluded state
        }
    });
}

function toggleGitignoreHandler(value) {
    useGitignore.value = value;
    addLog(
        `.gitignore usage changed to: ${value}. updating tree and watcher...`,
        "info",
        "bottom"
    );
    SetUseGitignore(value)
        .then(() =>
            addLog(`watchman instructed to use .gitignore: ${value}`, "debug")
        )
        .catch((err) =>
            addLog(`error setting usegitignore in backend: ${err}`, "error")
        );
    // context regeneration is handled by the watch on [filetree, usegitignore, usecustomignore]
    // which calls updateallnodesexcludedstate and debouncedtriggershotguncontextgeneration.
}

function toggleCustomIgnoreHandler(value) {
    useCustomIgnore.value = value;
    addLog(
        `custom ignore rules usage changed to: ${value}. updating tree and watcher...`,
        "info",
        "bottom"
    );
    SetUseCustomIgnore(value)
        .then(() =>
            addLog(
                `watchman instructed to use custom ignores: ${value}`,
                "debug"
            )
        )
        .catch((err) =>
            addLog(`error setting usecustomignore in backend: ${err}`, "error")
        );
}

function selectAllFiles() {
    if (!fileTree.value || fileTree.value.length === 0) return;

    function selectNodesRecursive(nodes) {
        if (!nodes || nodes.length === 0) return;

        nodes.forEach(node => {
            // Set node as included (not excluded)
            node.excluded = false;
            manuallyToggledNodes.set(node.relPath, false);

            // Process children recursively
            if (node.children && node.children.length > 0) {
                selectNodesRecursive(node.children);
            }
        });
    }

    selectNodesRecursive(fileTree.value);
    addLog("selected all files in the tree", "success", "bottom");

    // force context regeneration with a slight delay to ensure UI updates first
    setTimeout(() => {
        shotgunPromptContext.value = ""; // Clear current context to ensure fresh generation
        debouncedTriggerShotgunContextGeneration();
    }, 100);
}

function deselectAllFiles() {
    if (!fileTree.value || fileTree.value.length === 0) return;

    function deselectNodesRecursive(nodes) {
        if (!nodes || nodes.length === 0) return;

        nodes.forEach(node => {
            // Set node as excluded
            node.excluded = true;
            manuallyToggledNodes.set(node.relPath, true);

            // Process children recursively
            if (node.children && node.children.length > 0) {
                deselectNodesRecursive(node.children);
            }
        });
    }

    deselectNodesRecursive(fileTree.value);
    addLog("deselected all files in the tree", "success", "bottom");

    // clear existing context and prevent new generation
    shotgunPromptContext.value = "";
    isGeneratingContext.value = false;
    // cancel any pending debounced generation
    clearTimeout(debounceTimer);
    addLog("context generation canceled - all files are excluded", "info", "bottom");
}

function resetFileSelections() {
    // clear manual toggles to revert to default ignore states
    manuallyToggledNodes.clear();
    updateAllNodesExcludedState(fileTree.value);

    addLog("reset file selections to default ignore rules", "success", "bottom");

    // regenerate context based on new selection state
    shotgunPromptContext.value = "";
    debouncedTriggerShotgunContextGeneration();
}

function debouncedTriggerShotgunContextGeneration() {
    // skip generation when no files are visually included
    function anyNodesIncluded(nodes) {
        if (!nodes) return false;
        for (const n of nodes) {
            if (!n.excluded) return true;
            if (n.children && n.children.length > 0 && anyNodesIncluded(n.children)) {
                return true;
            }
        }
        return false;
    }

    if (!anyNodesIncluded(fileTree.value)) {
        addLog("all files are currently excluded, skipping context generation", "info", "bottom");
        shotgunPromptContext.value = "";
        generationProgressData.value = { current: 0, total: 0 };
        isGeneratingContext.value = false;
        clearTimeout(debounceTimer);
        return;
    }
    if (!projectRoot.value) {
        // clear context and stop loading if no project root
        shotgunPromptContext.value = ""; // clear previous context
        generationProgressData.value = { current: 0, total: 0 }; // reset progress
        // isgeneratingcontext will be set to false by the return or by the timeout if it runs
        isGeneratingContext.value = false;
        return;
    }

    if (isFileTreeLoading.value) {
        addLog(
            "debounced trigger skipped: file tree is loading.",
            "debug",
            "bottom"
        );
        isGeneratingContext.value = false;
        return;
    }

    if (!isGeneratingContext.value)
        nextTick(() => (isGeneratingContext.value = true));

    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
        if (!projectRoot.value) {
            isGeneratingContext.value = false;
            return;
        }
        if (isFileTreeLoading.value) {
            addLog(
                "debounced execution skipped: file tree became loading.",
                "debug",
                "bottom"
            );
            isGeneratingContext.value = false;
            return;
        }

        addLog(
            "debounced trigger: requesting shotgun context generation...",
            "info"
        );

        updateAllNodesExcludedState(fileTree.value);
        generationProgressData.value = { current: 0, total: 0 }; // reset progress before new request

        const excludedPathsArray = [];

        // helper to determine if a node has any visually included (checkbox checked) descendants
        function hasVisuallyIncludedDescendant(node) {
            if (!node.isDir || !node.children || node.children.length === 0) {
                return false;
            }
            for (const child of node.children) {
                if (!child.excluded) {
                    // if child itself is visually included (checkbox is checked)
                    return true;
                }
                if (hasVisuallyIncludedDescendant(child)) {
                    // or if any of its descendants are
                    return true;
                }
            }
            return false;
        }

        function collectTrulyExcludedPaths(nodes) {
            if (!nodes) return;
            nodes.forEach((node) => {
                // a node is truly excluded if its checkbox is unchecked (node.excluded is true)
                // and it does not have any descendant that is checked (visually included).
                if (node.excluded && !hasVisuallyIncludedDescendant(node)) {
                    excludedPathsArray.push(node.relPath);
                    // if a node is truly excluded, its children are implicitly excluded from generation,
                    // so no need to recurse further for collecting excluded paths under this node.
                } else {
                    // if the node is visually included or it's visually excluded but has an included descendant
                    // (meaning this node's path needs to be in the tree structure for its descendant),
                    // then we must check its children for their own exclusion status.
                    if (node.children && node.children.length > 0) {
                        collectTrulyExcludedPaths(node.children);
                    }
                }
            });
        }
        collectTrulyExcludedPaths(fileTree.value);

        RequestShotgunContextGeneration(projectRoot.value, excludedPathsArray)
            .catch((err) => {
                const errorMsg =
                    "error calling requestshotguncontextgeneration: " +
                    (err.message || err);
                addLog(errorMsg, "error");
                shotgunPromptContext.value = "error: " + errorMsg;
            })
            .finally(() => {
                // isgeneratingcontext.value = false;
            });
    }, 750);
}

function navigateToStep(stepId) {
    const targetStep = steps.value.find((s) => s.id === stepId);
    if (!targetStep) return;

    // if trying to stay on the same step, simply return
    if (stepId === currentStep.value) {
        currentStep.value = stepId;
        return;
    }

    // navigating backwards
    if (stepId < currentStep.value) {
        // whenever we move to an earlier step, subsequent steps should be marked as incomplete
        // this avoids inconsistent states where later steps remain completed after revisions
        for (let i = 0; i < steps.value.length; i++) {
            if (steps.value[i].id > stepId) {
                steps.value[i].completed = false;
            }
        }

        currentStep.value = stepId;
        return;
    }

    // navigating forwards to an already completed step
    if (targetStep.completed) {
        currentStep.value = stepId;
        return;
    }

    // navigating forwards to the first incomplete step is allowed
    const firstUncompletedStep = steps.value.find((s) => !s.completed);
    if (firstUncompletedStep && stepId === firstUncompletedStep.id) {
        currentStep.value = stepId;
    } else {
        const requiredStepId = firstUncompletedStep ? firstUncompletedStep.id : 'unknown';
        addLog(
            `cannot navigate to step ${stepId} yet. please complete step ${requiredStepId}.`,
            "warn"
        );
    }
}

function handleComposedPromptUpdate(prompt) {
    composedLlmPrompt.value = prompt;
    finalPrompt.value = prompt;
    addLog(
        `mainlayout: composed llm prompt updated (${prompt.length} chars).`,
        "debug",
        "bottom"
    );
    // logic to mark step 2 as complete can go here
    if (currentStep.value === 2 && prompt && steps.value[0].completed) {
        const step2 = steps.value.find((s) => s.id === 2);
        if (step2 && !step2.completed) {
            step2.completed = true;
            addLog(
                "step 2: prompt composed. ready to proceed to step 3.",
                "success",
                "bottom"
            );
        }
    }
}

async function handleStepAction(actionName, payload) {
    addLog(
        `action: ${actionName} triggered from step ${currentStep.value}.`,
        "info",
        "bottom"
    );
    if (payload && actionName === "composePrompt") {
        addLog(`prompt for diff: "${payload.prompt}"`, "info", "bottom");
    }

    const currentStepObj = steps.value.find((s) => s.id === currentStep.value);

    switch (actionName) {
        case "executePrompt":
            if (!composedLlmPrompt.value) {
                addLog(
                    "cannot execute prompt: prompt from step 2 is empty.",
                    "warn",
                    "both"
                );
                return;
            }
            addLog(
                `simulating backend: executing prompt (llm call)... \nprompt preview (first 100 chars): "${composedLlmPrompt.value.substring(
                    0,
                    100
                )}..."`,
                "info",
                "step"
            );
            // here, you would actually send composedllmprompt.value to an llm
            await new Promise((resolve) => setTimeout(resolve, 1000));
            addLog(
                "backend: llm call simulated. (mocked response/diff would be processed here).",
                "info",
                "step"
            );
            if (currentStepObj) currentStepObj.completed = true;
            // for now, just navigate to step 4, as step 3's "execution" is conceptual.
            // in a real app, step 3 might display llm output before proceeding.
            navigateToStep(4);
            break;
        case "executePromptAndSplitDiff": // handle the actual splitting action
            if (!payload || !payload.gitDiff || payload.lineLimit <= 0) {
                addLog(
                    "invalid payload for splitting diff.",
                    "error",
                    "bottom"
                );
                return;
            }
            addLog(
                `splitting diff (approx ${payload.lineLimit} lines per split)...`,
                "info",
                "bottom"
            );
            isLoadingSplitDiffs.value = true;
            splitDiffs.value = []; // clear previous splits
            shotgunGitDiff.value = payload.gitDiff;
            splitLineLimitValue.value = payload.lineLimit; // store the line limit
            try {
                const result = await SplitShotgunDiff(
                    payload.gitDiff,
                    payload.lineLimit
                );
                splitDiffs.value = result;
                addLog(
                    `diff split into ${result.length} parts.`,
                    "success",
                    "bottom"
                );

                if (currentStepObj) currentStepObj.completed = true;
                navigateToStep(4);
            } catch (err) {
                const errorMsg = `error splitting diff: ${err.message || err}`;
                addLog(errorMsg, "error", "bottom");
            } finally {
                isLoadingSplitDiffs.value = false;
            }
            break;
        case "applySelectedPatches":
        case "applyAllPatches":
            addLog(
                `simulating backend: applying patches (${actionName})...`,
                "info",
                "bottom"
            );
            await new Promise((resolve) => setTimeout(resolve, 1000));
            addLog(
                "backend: patches applied. process complete!",
                "info",
                "bottom"
            );
            if (currentStepObj) currentStepObj.completed = true;
            break;
        case "finishSplitting":
            addLog("finished with split diffs.", "info", "bottom");
            if (currentStepObj) currentStepObj.completed = true;
            break;
        default:
            addLog(`unknown action: ${actionName}`, "error", "bottom");
    }
}

// add global keyboard shortcut handler for navigating steps
function handleGlobalKeydown(event) {
    // ignore if the user is typing in an editable element
    const tag = event.target?.tagName?.toLowerCase() || "";
    const isEditable =
        event.target?.isContentEditable ||
        tag === "input" ||
        tag === "textarea" ||
        tag === "select";
    if (isEditable) return;

    // support ctrl (windows/linux) and meta (mac) + number 1-4
    if ((event.ctrlKey || event.metaKey) && !event.shiftKey && !event.altKey) {
        const key = event.key;
        if (key === "1" || key === "2" || key === "3" || key === "4") {
            event.preventDefault();
            navigateToStep(parseInt(key));
        }
    }
}

onMounted(() => {
    EventsOn("shotgunContextGenerated", (output) => {
        addLog(
            "wails event: shotguncontextgenerated received",
            "debug",
            "bottom"
        );
        shotgunPromptContext.value = output;
        isGeneratingContext.value = false;
        addLog(`shotgun context updated (${output.length} chars).`, "success");
        const step1 = steps.value.find((s) => s.id === 1);
        if (step1 && !step1.completed) {
            step1.completed = true;
        }
        if (
            currentStep.value === 1 &&
            centralPanelRef.value?.updateStep2ShotgunContext
        ) {
            centralPanelRef.value.updateStep2ShotgunContext(output);
        }
        checkAndProcessPendingFileTreeReload(); // check after context generation
    });

    EventsOn("shotgunContextError", (errorMsg) => {
        addLog(
            `wails event: shotguncontexterror received: ${errorMsg}`,
            "debug",
            "bottom"
        );
        shotgunPromptContext.value = "error: " + errorMsg;
        isGeneratingContext.value = false;
        addLog(`error generating context: ${errorMsg}`, "error");
        checkAndProcessPendingFileTreeReload(); // check after context generation error
    });

    EventsOn("shotgunContextGenerationProgress", (progress) => {
        // console.log("fe: progress event:", progress); // for debugging in browser console
        generationProgressData.value = progress;
    });

    // get platform information
    (async () => {
        try {
            const envInfo = await Environment();
            platform.value = envInfo.platform;
            addLog(`platform detected: ${platform.value}`, "debug");
        } catch (err) {
            addLog(`error getting platform: ${err}`, "error");
            // platform.value remains 'unknown' as fallback
        }
    })();

    unlistenProjectFilesChanged = EventsOn(
        "projectFilesChanged",
        (changedRootDir) => {
            if (changedRootDir !== projectRoot.value) {
                addLog(
                    `watchman: ignoring event for ${changedRootDir}, current root is ${projectRoot.value}`,
                    "debug"
                );
                return;
            }
            addLog(
                `watchman: event "projectfileschanged" received for ${changedRootDir}.`,
                "debug"
            );
            if (isFileTreeLoading.value || isGeneratingContext.value) {
                projectFilesChangedPendingReload.value = true;
                addLog(
                    "watchman: file change detected, reload queued as system is busy.",
                    "info"
                );
            } else {
                addLog(
                    "watchman: file change detected, reloading tree immediately.",
                    "info"
                );
                loadFileTree(projectRoot.value); // this will set isfiletreeloading = true
                // debouncedtriggershotguncontextgeneration will be called by the watcher on filetree if projectroot is set
            }
        }
    );

    // register the global keydown listener
    window.addEventListener("keydown", handleGlobalKeydown);

    // listen for the backend event that signals an initial folder should be opened automatically
    unlistenAutoOpenFolder = EventsOn("auto-open-folder", async (folderPath) => {
        if (typeof folderPath !== "string" || !folderPath) return;
        addLog(`auto-open-folder event received for path: ${folderPath}`, "info", "bottom");
        await openFolderProgrammatically(folderPath);
    });
});

onBeforeUnmount(async () => {
    clearTimeout(debounceTimer);
    if (projectRoot.value) {
        await StopFileWatcher().catch((err) =>
            console.error("error stopping file watcher on unmount:", err)
        );
        addLog(
            `file watcher stopped on component unmount for ${projectRoot.value}`,
            "debug"
        );
    }
    if (unlistenProjectFilesChanged) {
        unlistenProjectFilesChanged();
    }
    if (unlistenAutoOpenFolder) {
        unlistenAutoOpenFolder();
    }
    // remember to unlisten other events if they return unlistener functions

    // cleanup the global keydown listener
    window.removeEventListener("keydown", handleGlobalKeydown);
});

watch(
    [fileTree, useGitignore, useCustomIgnore],
    (
        [newFileTree, newUseGitignore, newUseCustomIgnore],
        [oldFileTree, oldUseGitignore, oldUseCustomIgnore]
    ) => {
        if (isFileTreeLoading.value) {
            addLog(
                "watcher triggered during file tree load, generation deferred.",
                "debug",
                "bottom"
            );
            return;
        }

        addLog(
            "watcher detected changes in filetree, usegitignore, or usecustomignore. re-evaluating context.",
            "debug",
            "bottom"
        );
        updateAllNodesExcludedState(fileTree.value);
        debouncedTriggerShotgunContextGeneration();
    },
    { deep: true }
);

watch(
    projectRoot,
    async (newRoot, oldRoot) => {
        if (oldRoot) {
            await StopFileWatcher().catch((err) =>
                addLog(`error stopping watcher for ${oldRoot}: ${err}`, "error")
            );
            addLog(`file watcher stopped for ${oldRoot}`, "debug");
        }
        if (newRoot) {
            // existing logic to loadfiletree, clear errors, etc., happens in selectprojectfolderhandler
            // which sets projectroot. here we just ensure the watcher starts for the new root.
            await StartFileWatcher(newRoot).catch((err) =>
                addLog(`error starting watcher for ${newRoot}: ${err}`, "error")
            );
            addLog(`file watcher started for ${newRoot}`, "debug");
        } else {
            // project root cleared, ensure watcher is stopped (already handled by oldroot check if it was set)
            fileTree.value = [];
            shotgunPromptContext.value = "";
            loadingError.value = "";
            manuallyToggledNodes.clear();
            isGeneratingContext.value = false; // reset generation state
            projectFilesChangedPendingReload.value = false; // reset pending reload
        }
    },
    { immediate: false }
); // 'immediate: false' to avoid running on initial undefined -> '' or '' -> initial value if set by default

// helper function to process pending reloads
function checkAndProcessPendingFileTreeReload() {
    if (
        projectFilesChangedPendingReload.value &&
        !isFileTreeLoading.value &&
        !isGeneratingContext.value
    ) {
        projectFilesChangedPendingReload.value = false;
        addLog("watchman: processing queued file tree reload.", "info");
        // it's important that loadfiletree correctly sets isfiletreeloading to true at its start
        // and that subsequent context generation is also handled.
        loadFileTree(projectRoot.value);
    }
}

function handleCustomRulesUpdated() {
    addLog("custom ignore rules updated by user. reloading file tree.", "info");
    if (projectRoot.value) {
        // this will call listfiles in go, which will use the new custom rules from app.settings.
        // the new tree will have updated iscustomignored flags.
        // the watch on filetree (and its subsequent call to debouncedtriggershotguncontextgeneration)
        // will then handle regenerating the context.
        loadFileTree(projectRoot.value);
    }
}

function handleUserTaskUpdate(val) {
    userTask.value = val;
}

function handleRulesContentUpdate(val) {
    rulesContent.value = val;
}

// add handlers for the new updates
function handleShotgunGitDiffUpdate(val) {
    shotgunGitDiff.value = val;
}

function handleSplitLineLimitUpdate(val) {
    splitLineLimitValue.value = val;
}

// programmatically open a folder path without showing the select-dialog.
async function openFolderProgrammatically(folderPath) {
    if (!folderPath) return;
    isFileTreeLoading.value = true;
    try {
        shotgunPromptContext.value = "";
        isGeneratingContext.value = false;
        projectRoot.value = folderPath;
        loadingError.value = "";
        manuallyToggledNodes.clear();
        fileTree.value = [];

        await loadFileTree(folderPath);

        splitDiffs.value = [];

        if (!isFileTreeLoading.value && projectRoot.value) {
            debouncedTriggerShotgunContextGeneration();
        }

        steps.value.forEach((s) => (s.completed = false));
        currentStep.value = 1;
        addLog(`project folder opened automatically: ${folderPath}`, "info", "bottom");
    } catch (err) {
        console.error("error opening folder programmatically:", err);
        const errorMsg = `failed to open folder: ${err.message || err}`;
        loadingError.value = errorMsg;
        addLog(errorMsg, "error", "bottom");
    } finally {
        isFileTreeLoading.value = false;
    }
}
</script>

<style scoped>
.flex-1 {
    min-height: 0;
}
</style>
