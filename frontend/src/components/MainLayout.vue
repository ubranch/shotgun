<template>
  <div class="flex flex-col h-screen bg-gray-100">
    <HorizontalStepper :current-step="currentStep" :steps="steps" @navigate="navigateToStep" />
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
        @toggle-exclude="toggleExcludeNode" />
      <CentralPanel :current-step="currentStep" 
                    :shotgun-prompt-context="shotgunPromptContext"
                    :generation-progress="generationProgressData"
                    :is-generating-context="isGeneratingContext"
                    :project-root="projectRoot" 
                    :platform="platform"
                    @step-action="handleStepAction"
                    @update-composed-prompt="handleComposedPromptUpdate" 
                    ref="centralPanelRef" />
    </div>
    <div 
      @mousedown="startResize"
      class="w-full h-2 bg-gray-300 hover:bg-gray-400 cursor-row-resize select-none"
      title="Resize console height"
    >
    </div>
    <BottomConsole :log-messages="logMessages" :height="consoleHeight" ref="bottomConsoleRef" />
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted, onBeforeUnmount, nextTick } from 'vue';
import HorizontalStepper from './HorizontalStepper.vue';
import LeftSidebar from './LeftSidebar.vue';
import CentralPanel from './CentralPanel.vue';
import BottomConsole from './BottomConsole.vue';
import { ListFiles, RequestShotgunContextGeneration, SelectDirectory as SelectDirectoryGo } from '../../wailsjs/go/main/App';
import { EventsOn, Environment } from '../../wailsjs/runtime/runtime';

const currentStep = ref(1);
const steps = ref([
  { id: 1, title: 'Prepare Context', completed: false, description: 'Select project folder, review files, and generate the initial project context for the LLM.' },
  { id: 2, title: 'Compose Prompt', completed: false, description: 'Provide a prompt to the LLM based on the project context to generate a code diff.' },
  { id: 3, title: 'Execute Prompt', completed: false, description: 'Review and simulate the execution of the generated diff to understand its impact.' },
  { id: 4, title: 'Apply Patch', completed: false, description: 'Apply the verified changes (patches) to your actual project files.' },
]);

const logMessages = ref([]);
const centralPanelRef = ref(null); 
const bottomConsoleRef = ref(null);
const MIN_CONSOLE_HEIGHT = 50;
const consoleHeight = ref(MIN_CONSOLE_HEIGHT); // Initial height in pixels

function addLog(message, type = 'info', targetConsole = 'bottom') {
  const logEntry = {
    message,
    type,
    timestamp: new Date().toLocaleTimeString()
  };

  if (targetConsole === 'bottom' || targetConsole === 'both') {
    logMessages.value.push(logEntry);
  }
  if (targetConsole === 'step' || targetConsole === 'both') {
    if (centralPanelRef.value && currentStep.value === 3 && centralPanelRef.value.addLogToStep3Console) {
      centralPanelRef.value.addLogToStep3Console(message, type);
    }
  }
}

const projectRoot = ref('');
const fileTree = ref([]);
const shotgunPromptContext = ref('');
const loadingError = ref('');
const useGitignore = ref(true);
const useCustomIgnore = ref(true);
const manuallyToggledNodes = reactive(new Map());
const isGeneratingContext = ref(false);
const generationProgressData = ref({ current: 0, total: 0 });
const isFileTreeLoading = ref(false);
const composedLlmPrompt = ref(''); // To store the prompt from Step 2
const platform = ref('unknown'); // To store OS platform (e.g., 'darwin', 'windows', 'linux')
let debounceTimer = null;

async function selectProjectFolderHandler() {
  isFileTreeLoading.value = true;
  try {
    shotgunPromptContext.value = '';
    isGeneratingContext.value = false;
    const selectedDir = await SelectDirectoryGo(); 
    if (selectedDir) {
      projectRoot.value = selectedDir;
      loadingError.value = '';
      manuallyToggledNodes.clear();
      fileTree.value = [];
      
      await loadFileTree(selectedDir);

      if (!isFileTreeLoading.value && projectRoot.value) {
         debouncedTriggerShotgunContextGeneration();
      }

      steps.value.forEach(s => s.completed = false);
      currentStep.value = 1;
      addLog(`Project folder selected: ${selectedDir}`, 'info', 'bottom');
    } else {
      isFileTreeLoading.value = false;
    }
  } catch (err) {
    console.error("Error selecting directory:", err);
    const errorMsg = "Failed to select directory: " + (err.message || err);
    loadingError.value = errorMsg;
    addLog(errorMsg, 'error', 'bottom');
    isFileTreeLoading.value = false;
  }
}

async function loadFileTree(dirPath) {
  isFileTreeLoading.value = true;
  loadingError.value = '';
  addLog(`Loading file tree for: ${dirPath}`, 'info', 'bottom');
  try {
    const treeData = await ListFiles(dirPath);
    fileTree.value = mapDataToTreeRecursive(treeData, null);
    addLog(`File tree loaded successfully. Root items: ${fileTree.value.length}`, 'info', 'bottom');
  } catch (err) {
    console.error("Error listing files:", err);
    const errorMsg = "Failed to load file tree: " + (err.message || err);
    loadingError.value = errorMsg;
    addLog(errorMsg, 'error', 'bottom');
    fileTree.value = [];
  } finally {
    isFileTreeLoading.value = false;
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
  return nodes.map(node => {
    const isRootNode = parent === null;
    const reactiveNode = reactive({
      ...node,
      expanded: node.isDir ? isRootNode : undefined,
      parent: parent,
      children: [] 
    });
    reactiveNode.excluded = calculateNodeExcludedState(reactiveNode);

    if (node.children && node.children.length > 0) {
      reactiveNode.children = mapDataToTreeRecursive(node.children, reactiveNode);
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
    if (current.excluded) { // current.excluded reflects its visual/checkbox state
      return true;
    }
    current = current.parent;
  }
  return false;
}

function toggleExcludeNode(nodeToToggle) {
  // If the node is under an unselected parent and is currently unselected itself (nodeToToggle.excluded is true),
  // the first click should select it (set nodeToToggle.excluded to false).
  if (isAnyParentVisuallyExcluded(nodeToToggle) && nodeToToggle.excluded) {
    nodeToToggle.excluded = false;
  } else {
    // Otherwise, normal toggle behavior.
    nodeToToggle.excluded = !nodeToToggle.excluded;
  }
  manuallyToggledNodes.set(nodeToToggle.relPath, nodeToToggle.excluded);
  addLog(`Toggled exclusion for ${nodeToToggle.name} to ${nodeToToggle.excluded}`, 'info', 'bottom');
}

function updateAllNodesExcludedState(nodesToUpdate) { // This is the public-facing function
  // It calls the recursive helper, starting with parentIsVisuallyExcluded = false for root nodes.
  _updateAllNodesExcludedStateRecursive(nodesToUpdate, false);
}

function _updateAllNodesExcludedStateRecursive(nodesToUpdate, parentIsVisuallyExcluded) {
   if (!nodesToUpdate || nodesToUpdate.length === 0) return;
   nodesToUpdate.forEach(node => {
    const manualToggle = manuallyToggledNodes.get(node.relPath);
    let isExcludedByRule = false;
    if (useGitignore.value && node.isGitignored) isExcludedByRule = true;
    if (useCustomIgnore.value && node.isCustomIgnored) isExcludedByRule = true;

    if (manualToggle !== undefined) {
      // If there's a manual toggle, it dictates the state.
      node.excluded = manualToggle;
    } else {
      // If not manually toggled, it's excluded if a rule matches OR if its parent is visually excluded.
      // This establishes the default inherited exclusion for visual purposes.
      node.excluded = isExcludedByRule || parentIsVisuallyExcluded;
    }

     if (node.children && node.children.length > 0) {
      _updateAllNodesExcludedStateRecursive(node.children, node.excluded); // Pass current node's new visual excluded state
     }
   });
 }

function toggleGitignoreHandler(value) {
  useGitignore.value = value;
  addLog(`.gitignore usage changed to: ${value}. Updating tree...`, 'info', 'bottom');
}

function toggleCustomIgnoreHandler(value) {
  useCustomIgnore.value = value;
  addLog(`Custom ignore rules usage changed to: ${value}. Updating tree...`, 'info', 'bottom');
}

function debouncedTriggerShotgunContextGeneration() {
  if (!projectRoot.value) {
    isGeneratingContext.value = false;
    return;
  }

  if (isFileTreeLoading.value) {
    addLog("Debounced trigger skipped: file tree is loading.", 'debug', 'bottom');
    isGeneratingContext.value = false;
    return;
  }

  if (!isGeneratingContext.value) nextTick(() => isGeneratingContext.value = true);

  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    if (!projectRoot.value) { 
        isGeneratingContext.value = false;
        return;
    }
    if (isFileTreeLoading.value) {
        addLog("Debounced execution skipped: file tree became loading.", 'debug', 'bottom');
        isGeneratingContext.value = false;
        return;
    }

    addLog("Debounced trigger: Requesting shotgun context generation...", 'info');
    
    updateAllNodesExcludedState(fileTree.value);
    generationProgressData.value = { current: 0, total: 0 }; // Reset progress before new request

    const excludedPathsArray = [];
    
    // Helper to determine if a node has any visually included (checkbox checked) descendants
    function hasVisuallyIncludedDescendant(node) {
      if (!node.isDir || !node.children || node.children.length === 0) {
        return false;
      }
      for (const child of node.children) {
        if (!child.excluded) { // If child itself is visually included (checkbox is checked)
          return true;
        }
        if (hasVisuallyIncludedDescendant(child)) { // Or if any of its descendants are
          return true;
        }
      }
      return false;
    }

    function collectTrulyExcludedPaths(nodes) {
       if (!nodes) return;
       nodes.forEach(node => {
        // A node is TRULY excluded if its checkbox is unchecked (node.excluded is true)
        // AND it does not have any descendant that is checked (visually included).
        if (node.excluded && !hasVisuallyIncludedDescendant(node)) {
          excludedPathsArray.push(node.relPath);
          // If a node is truly excluded, its children are implicitly excluded from generation,
          // so no need to recurse further for collecting excluded paths under this node.
        } else {
          // If the node is visually included OR it's visually excluded but has an included descendant
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
       .catch(err => {
        const errorMsg = "Error calling RequestShotgunContextGeneration: " + (err.message || err);
        addLog(errorMsg, 'error');
        shotgunPromptContext.value = "Error: " + errorMsg; 
      })
      .finally(() => {
         // isGeneratingContext.value = false;
      });
  }, 750); 
}

function navigateToStep(stepId) {
  const targetStep = steps.value.find(s => s.id === stepId);
  if (!targetStep) return;

  if (targetStep.completed || stepId === currentStep.value) {
    currentStep.value = stepId;
    return;
  }

  const firstUncompletedStep = steps.value.find(s => !s.completed);
  if (!firstUncompletedStep || stepId === firstUncompletedStep.id) {
    currentStep.value = stepId;
  } else {
    addLog(`Cannot navigate to step ${stepId} yet. Please complete step ${firstUncompletedStep.id}.`, 'warn');
  }
}

function handleComposedPromptUpdate(prompt) {
  composedLlmPrompt.value = prompt;
  addLog(`MainLayout: Composed LLM prompt updated (${prompt.length} chars).`, 'debug', 'bottom');
  // Logic to mark step 2 as complete can go here
  if (currentStep.value === 2 && prompt && steps.value[0].completed) {
    const step2 = steps.value.find(s => s.id === 2);
    if (step2 && !step2.completed) {
      step2.completed = true;
      addLog("Step 2: Prompt composed. Ready to proceed to Step 3.", "success", "bottom");
    }
  }
}

async function handleStepAction(actionName, payload) {
  addLog(`Action: ${actionName} triggered from step ${currentStep.value}.`, 'info', 'bottom');
  if (payload && actionName === 'composePrompt') {
    addLog(`Prompt for diff: "${payload.prompt}"`, 'info', 'bottom');
  }

  const currentStepObj = steps.value.find(s => s.id === currentStep.value);

  switch (actionName) {
    case 'executePrompt':
      if (!composedLlmPrompt.value) {
        addLog("Cannot execute prompt: Prompt from Step 2 is empty.", 'warn', 'both');
        return;
      }
      addLog(`Simulating backend: Executing prompt (LLM call)... \nPrompt Preview (first 100 chars): "${composedLlmPrompt.value.substring(0,100)}..."`, 'info', 'step');
      // Here, you would actually send composedLlmPrompt.value to an LLM
      await new Promise(resolve => setTimeout(resolve, 1000));
      addLog('Backend: LLM call simulated. (Mocked response/diff would be processed here).', 'info', 'step');
      if (currentStepObj) currentStepObj.completed = true;
      // For now, just navigate to Step 4, as Step 3's "execution" is conceptual.
      // In a real app, Step 3 might display LLM output before proceeding.
      navigateToStep(4); 
      break;
    case 'applySelectedPatches':
    case 'applyAllPatches':
      addLog(`Simulating backend: Applying patches (${actionName})...`, 'info', 'bottom');
      await new Promise(resolve => setTimeout(resolve, 1000));
      addLog('Backend: Patches applied. Process complete!', 'info', 'bottom');
      if (currentStepObj) currentStepObj.completed = true;
      break;
    default:
      addLog(`Unknown action: ${actionName}`, 'error', 'bottom');
  }
}

const isResizing = ref(false);

function startResize(event) {
  isResizing.value = true;
  document.addEventListener('mousemove', doResize);
  document.addEventListener('mouseup', stopResize);
  event.preventDefault(); 
}

function doResize(event) {
  if (!isResizing.value) return;
  const newHeight = window.innerHeight - event.clientY;
  const minHeight = MIN_CONSOLE_HEIGHT;
  const maxHeight = window.innerHeight * 0.7;
  consoleHeight.value = Math.max(minHeight, Math.min(newHeight, maxHeight));
}

function stopResize() {
  isResizing.value = false;
  document.removeEventListener('mousemove', doResize);
  document.removeEventListener('mouseup', stopResize);
}

onMounted(() => {
  EventsOn("shotgunContextGenerated", (output) => {
    addLog("Wails event: shotgunContextGenerated RECEIVED", 'debug', 'bottom');
    shotgunPromptContext.value = output;
    isGeneratingContext.value = false;
    addLog(`Shotgun context updated (${output.length} chars).`, 'success');
    const step1 = steps.value.find(s => s.id === 1);
    if (step1 && !step1.completed) {
        step1.completed = true;
    }
    if (currentStep.value === 1 && centralPanelRef.value?.updateStep2ShotgunContext) {
        centralPanelRef.value.updateStep2ShotgunContext(output);
    }
  });

  EventsOn("shotgunContextError", (errorMsg) => {
    addLog(`Wails event: shotgunContextError RECEIVED: ${errorMsg}`, 'debug', 'bottom');
    shotgunPromptContext.value = "Error: " + errorMsg;
    isGeneratingContext.value = false;
    addLog(`Error generating context: ${errorMsg}`, 'error');
  });

  EventsOn("shotgunContextGenerationProgress", (progress) => {
    // console.log("FE: Progress event:", progress); // For debugging in Browser console
    generationProgressData.value = progress;
  });

  // Get platform information
  (async () => {
    try {
      const envInfo = await Environment();
      platform.value = envInfo.platform;
      addLog(`Platform detected: ${platform.value}`, 'debug');
    } catch (err) {
      addLog(`Error getting platform: ${err}`, 'error');
      // platform.value remains 'unknown' as fallback
    }
  })();
});

onBeforeUnmount(() => {
  document.removeEventListener('mousemove', doResize);
  document.removeEventListener('mouseup', stopResize);
  clearTimeout(debounceTimer);
});

watch([fileTree, useGitignore, useCustomIgnore], ([newFileTree, newUseGitignore, newUseCustomIgnore], [oldFileTree, oldUseGitignore, oldUseCustomIgnore]) => {
  if (isFileTreeLoading.value) {
    addLog("Watcher triggered during file tree load, generation deferred.", 'debug', 'bottom');
    return;
  }
  
  addLog("Watcher detected changes in fileTree, useGitignore, or useCustomIgnore. Re-evaluating context.", 'debug', 'bottom');
  updateAllNodesExcludedState(fileTree.value);
  debouncedTriggerShotgunContextGeneration();
}, { deep: true });

</script>

<style scoped>
.flex-1 {
  min-height: 0;
}
</style> 