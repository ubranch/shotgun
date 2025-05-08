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
        :loading-error="loadingError"
        @navigate="navigateToStep"
        @select-folder="selectProjectFolderHandler"
        @toggle-gitignore="toggleGitignoreHandler"
        @toggle-exclude="toggleExcludeNode" />
      <CentralPanel :current-step="currentStep" 
                    :shotgun-prompt-context="shotgunPromptContext" 
                    :step1-context-generation-attempted="step1ContextGenerationAttempted" 
                    :project-root="projectRoot" 
                    @step-action="handleStepAction" 
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
import { ref, reactive, watch, onMounted, onBeforeUnmount } from 'vue';
import HorizontalStepper from './HorizontalStepper.vue';
import LeftSidebar from './LeftSidebar.vue';
import CentralPanel from './CentralPanel.vue';
import BottomConsole from './BottomConsole.vue';

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

import { ListFiles, GenerateShotgunOutput, SelectDirectory as SelectDirectoryGo } from '../../wailsjs/go/main/App';

const projectRoot = ref('');
const fileTree = ref([]);
const shotgunPromptContext = ref('');
const loadingError = ref('');
const useGitignore = ref(true);
const manuallyToggledNodes = reactive(new Map());
const copyStatusText = ref('');
const step1ContextGenerationAttempted = ref(false);

async function selectProjectFolderHandler() {
  try {
    const selectedDir = await SelectDirectoryGo(); 
    if (selectedDir) {
      projectRoot.value = selectedDir;
      loadingError.value = '';
      manuallyToggledNodes.clear();
      shotgunPromptContext.value = '';
      fileTree.value = [];
      await loadFileTree(selectedDir);
      step1ContextGenerationAttempted.value = false;
      steps.value.forEach(s => s.completed = false);
      currentStep.value = 1;
      addLog(`Project folder selected: ${selectedDir}`, 'info', 'bottom');
    }
  } catch (err) {
    console.error("Error selecting directory:", err);
    const errorMsg = "Failed to select directory: " + (err.message || err);
    loadingError.value = errorMsg;
    addLog(errorMsg, 'error', 'bottom');
  }
}

async function loadFileTree(dirPath) {
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
  }
}

function calculateNodeExcludedState(node) {
  const manualToggle = manuallyToggledNodes.get(node.relPath);
  if (manualToggle !== undefined) return manualToggle;
  if (useGitignore.value && node.isGitignored) return true;
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

function toggleExcludeNode(node) {
  node.excluded = !node.excluded;
  manuallyToggledNodes.set(node.relPath, node.excluded);
  fileTree.value = [...fileTree.value];
  addLog(`Toggled exclusion for ${node.name} to ${node.excluded}`, 'info', 'bottom');
}

function updateAllNodesExcludedState(nodesToUpdate) {
  if (!nodesToUpdate) return;
  nodesToUpdate.forEach(node => {
    node.excluded = calculateNodeExcludedState(node);
    if (node.children && node.children.length > 0) {
      updateAllNodesExcludedState(node.children);
    }
  });
}

watch(useGitignore, (newValue) => {
  addLog(`.gitignore usage changed to: ${newValue}. Updating tree...`, 'info', 'bottom');
  updateAllNodesExcludedState(fileTree.value);
});

function toggleGitignoreHandler(value) {
  useGitignore.value = value;
}

async function generateShotgunPromptContext() {
  if (!projectRoot.value) {
    addLog("Cannot generate context: No project root selected.", 'warn', 'bottom');
    return "";
  }
  addLog("Generating Shotgun prompt context...", 'info', 'both');
  copyStatusText.value = ''; 

  const excludedPathsArray = [];
  function collectExcluded(nodes) {
    if (!nodes) return;
    nodes.forEach(node => {
      if (node.excluded) excludedPathsArray.push(node.relPath);
      if (node.children) collectExcluded(node.children);
    });
  }
  collectExcluded(fileTree.value);

  try {
    const result = await GenerateShotgunOutput(projectRoot.value, excludedPathsArray);
    shotgunPromptContext.value = result;
    addLog(`Shotgun prompt context generated (${result.length} characters).`, 'info', 'both');
    step1ContextGenerationAttempted.value = true;
    return result;
  } catch (err) {
    console.error("Error generating shotgun output:", err);
    const errorMsg = "Error generating context: " + (err.message || err);
    shotgunPromptContext.value = errorMsg;
    addLog(errorMsg, 'error', 'both');
    step1ContextGenerationAttempted.value = true;
    return errorMsg;
  }
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

async function handleStepAction(actionName, payload) {
  addLog(`Action: ${actionName} triggered from step ${currentStep.value}.`, 'info', 'bottom');
  if (payload && actionName === 'composePrompt') {
    addLog(`Prompt for diff: "${payload.prompt}"`, 'info', 'bottom');
  } else if (payload) {
    addLog(`Payload received for ${actionName}.`, 'info', 'bottom');
  }

  const currentStepObj = steps.value.find(s => s.id === currentStep.value);

  switch (actionName) {
    case 'prepareContext':
      if (!projectRoot.value) {
        addLog("Please select a project folder first.", 'warn', 'bottom');
        return;
      }
      addLog('Preparing project context...', 'info', 'bottom');
      await generateShotgunPromptContext();
      
      if (currentStepObj) currentStepObj.completed = true;
      step1ContextGenerationAttempted.value = true;
      if (centralPanelRef.value?.updateStep2ShotgunContext && shotgunPromptContext.value) {
         centralPanelRef.value.updateStep2ShotgunContext(shotgunPromptContext.value);
      }
      // navigateToStep(2); // Removed to prevent auto-navigation
      break;
    case 'composePrompt':
      if (!shotgunPromptContext.value) {
        addLog("Project context not generated. Please complete Step 1.", 'warn', 'both');
        return;
      }
      addLog(`Simulating backend: Generating diff...`, 'info', 'both');
      await new Promise(resolve => setTimeout(resolve, 1500));
      const mockDiff = `--- a/file1.txt\n+++ b/file1.txt\n@@ -1,1 +1,1 @@\n-Hello\n+World`;
      if (centralPanelRef.value?.updateStep2DiffOutput) centralPanelRef.value.updateStep2DiffOutput(mockDiff);
      addLog(`Backend: Diff generated (${mockDiff.length} characters).`, 'info', 'both');
      if (currentStepObj) currentStepObj.completed = true;
      navigateToStep(3);
      break;
    case 'executePrompt':
      addLog('Simulating backend: Executing diff... (mocked prompt execution)', 'info', 'step');
      await new Promise(resolve => setTimeout(resolve, 1000));
      addLog('Backend: Diff execution simulated (mocked prompt execution complete).', 'info', 'step');
      if (currentStepObj) currentStepObj.completed = true;
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

// Resize logic
const isResizing = ref(false);

function startResize(event) {
  isResizing.value = true;
  document.addEventListener('mousemove', doResize);
  document.addEventListener('mouseup', stopResize);
  // Prevent text selection during resize
  event.preventDefault(); 
}

function doResize(event) {
  if (!isResizing.value) return;
  // Calculate new height based on mouse position
  // window.innerHeight is the total viewport height
  // event.clientY is the mouse Y position relative to the viewport
  const newHeight = window.innerHeight - event.clientY;
  // Apply constraints (e.g., min/max height)
  const minHeight = MIN_CONSOLE_HEIGHT; // Min console height
  const maxHeight = window.innerHeight * 0.7; // Max 70% of viewport height
  consoleHeight.value = Math.max(minHeight, Math.min(newHeight, maxHeight));
}

function stopResize() {
  isResizing.value = false;
  document.removeEventListener('mousemove', doResize);
  document.removeEventListener('mouseup', stopResize);
}

onMounted(() => {
  // You might want to load a saved height from localStorage here
});

onBeforeUnmount(() => {
  // Clean up global event listeners if any were left (though stopResize should handle it)
  document.removeEventListener('mousemove', doResize);
  document.removeEventListener('mouseup', stopResize);
});
</script>

<style scoped>
/* Add any additional styles if needed */
.flex-1 {
  /* This ensures CentralPanel takes up remaining space */
  min-height: 0; /* Important for flex children to shrink properly */
}
</style> 