<template>
  <div class="app-container">
    <div class="left-panel">
      <button @click="selectProjectFolderHandler">Select Project Folder</button>
      <div v-if="projectRoot" class="project-path">Selected: {{ projectRoot }}</div>
      
      <div v-if="projectRoot" class="options-bar">
        <label>
          <input type="checkbox" v-model="useGitignore" />
          Use .gitignore rules
        </label>
      </div>

      <FileTree 
        v-if="fileTree.length" 
        :nodes="fileTree" 
        :project-root="projectRoot"
        @toggle-exclude="toggleExcludeNode"
      />
      <div v-else-if="projectRoot && !loadingError">Loading tree...</div>
      <div v-if="loadingError" class="error-message">{{ loadingError }}</div>

      <button v-if="projectRoot" @click="generateShotgun" class="shotgun-button">Shotgun</button>
    </div>
    <div class="right-panel">
      <pre class="output-display">{{ shotgunOutput }}</pre>
      <button v-if="shotgunOutput" @click="copyShotgunOutputToClipboard" class="copy-all-button">
        <span class="icon">📋</span> {{ copyButtonText }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch } from 'vue';
// Wails specific imports for Go methods. Adjust path if your main App struct is in a different Go package.
import { ListFiles, GenerateShotgunOutput } from '../wailsjs/go/main/App';
// For directory selection, Wails provides runtime functions.
// We'll create a Go wrapper for SelectDirectory for better consistency & control.
// Let's assume a Go method `SelectDirectory` is bound.
import { SelectDirectory as SelectDirectoryGo } from '../wailsjs/go/main/App'; // Assuming you add this wrapper in app.go
import FileTree from './components/FileTree.vue';

const projectRoot = ref('');
const fileTree = ref([]);
const shotgunOutput = ref('');
const loadingError = ref('');

const useGitignore = ref(true);
const manuallyToggledNodes = reactive(new Map()); // relPath => boolean (true if manually excluded, false if manually included)

const copyButtonText = ref('Copy All');

async function selectProjectFolderHandler() {
  try {
    // This Go method (SelectDirectoryGo) would internally call runtime.OpenDirectoryDialog
    const selectedDir = await SelectDirectoryGo(); 
    if (selectedDir) {
      projectRoot.value = selectedDir;
      loadingError.value = '';
      manuallyToggledNodes.clear(); // Clear manual toggles for new project
      shotgunOutput.value = ''; // Clear previous output when new folder selected
      copyButtonText.value = 'Copy All'; // Reset button text
      await loadFileTree(selectedDir);
    }
  } catch (err) {
    console.error("Error selecting directory:", err);
    loadingError.value = "Failed to select directory: " + (err.message || err);
  }
}

async function loadFileTree(dirPath) {
  try {
    const treeData = await ListFiles(dirPath);
    console.log("Raw treeData from backend:", JSON.parse(JSON.stringify(treeData)));
    fileTree.value = mapDataToTree(treeData, null);
  } catch (err) {
    console.error("Error listing files:", err);
    loadingError.value = "Failed to load file tree: " + (err.message || err);
    fileTree.value = [];
  }
}

function calculateNodeExcludedState(node, parent) {
  const manualToggle = manuallyToggledNodes.get(node.relPath);
  if (manualToggle !== undefined) {
    return manualToggle; // Manual toggle takes precedence
  }
  if (useGitignore.value && node.isGitignored) {
    return true; // Excluded by .gitignore
  }
  // If it's a directory and its parent is excluded (and not manually included), it should also be excluded
  // However, this is more complex as gitignore rules can re-include sub-items.
  // For now, let's rely on explicit gitignore flag for each item from backend and manual overrides.
  // If we wanted to inherit parent exclusion strictly when not using gitignore, that logic would go here.
  return false; // Default to not excluded
}

function mapDataToTree(nodes, parent) {
  if (!nodes) return [];
  return nodes.map(node => {
    const isRootNode = parent === null; // Check if current node is the root
    const reactiveNode = reactive({
      ...node, // Includes isGitignored from backend
      expanded: node.isDir ? (isRootNode ? true : false) : undefined, // Root expanded, others not
      // `excluded` state is now calculated based on gitignore and manual toggles
      // This will be set by calculateNodeExcludedState or by the watcher
      parent: parent,
      children: [] 
    });
    reactiveNode.excluded = calculateNodeExcludedState(reactiveNode, parent); // Set initial state

    if (node.children && node.children.length > 0) {
      reactiveNode.children = mapDataToTree(node.children, reactiveNode);
    }
    
    return reactiveNode;
  });
}

// Renamed from toggleExclude to avoid conflict with prop name in FileTree.vue if we pass it down directly
function toggleExcludeNode(node) {
  node.excluded = !node.excluded;
  manuallyToggledNodes.set(node.relPath, node.excluded);
  // Force Vue to recognize the change in the deeply nested property for rendering updates.
  // This is sometimes needed if direct modification isn't picked up for list re-rendering.
  fileTree.value = [...fileTree.value]; 
}

function updateAllNodesExcludedState(nodes) {
  if (!nodes) return;
  nodes.forEach(node => {
    node.excluded = calculateNodeExcludedState(node, node.parent);
    if (node.children && node.children.length > 0) {
      updateAllNodesExcludedState(node.children);
    }
  });
}

watch(useGitignore, () => {
  updateAllNodesExcludedState(fileTree.value);
});

async function generateShotgun() {
  if (!projectRoot.value) return;
  try {
    shotgunOutput.value = "Generating output...";
    copyButtonText.value = 'Copy All'; // Reset button text in case it was 'Copied!'
    const excludedPathsArray = [];
    function collectExcluded(nodes) {
      if (!nodes) return;
      nodes.forEach(node => {
        if (node.excluded) {
          excludedPathsArray.push(node.relPath);
        }
        if (node.children) {
          collectExcluded(node.children);
        }
      });
    }
    collectExcluded(fileTree.value);

    const result = await GenerateShotgunOutput(projectRoot.value, excludedPathsArray);
    shotgunOutput.value = result;
  } catch (err) {
    console.error("Error generating shotgun output:", err);
    shotgunOutput.value = "Error: " + (err.message || err);
  }
}

async function copyShotgunOutputToClipboard() {
  if (!shotgunOutput.value) return;
  try {
    await navigator.clipboard.writeText(shotgunOutput.value);
    copyButtonText.value = 'Copied!';
    setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000); // Reset text after 2 seconds
  } catch (err) {
    console.error('Failed to copy text: ', err);
    copyButtonText.value = 'Failed to copy';
     setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000);
  }
}

</script>

<style scoped>
.app-container {
  display: flex;
  height: 100vh;
  font-family: Arial, sans-serif;
}
.left-panel {
  width: 40%;
  padding: 10px;
  border-right: 1px solid #ccc;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.options-bar {
  padding: 8px 0;
  margin-bottom: 5px;
  border-bottom: 1px solid #eee;
}
.right-panel {
  width: 60%;
  padding: 10px;
  overflow-y: auto;
  display: flex; /* Added for flex layout */
  flex-direction: column; /* Children stack vertically */
}
.output-display {
  white-space: pre-wrap; 
  word-wrap: break-word; 
  font-family: monospace;
  font-size: 0.9em;
  flex-grow: 1; /* Takes available vertical space */
  overflow-y: auto; /* Scrollbar if content overflows */
  border: 1px solid #eee; /* Optional: to delineate the area */
  padding: 5px; /* Optional: internal padding */
  margin-bottom: 5px; /* Space before the copy button */
}
.copy-all-button {
  padding: 8px 15px;
  cursor: pointer;
  align-self: flex-start; /* Align button to the start of the cross axis (left in a column flex) */
  margin-top: 5px; /* Added margin for spacing if .output-display margin-bottom is removed */
}
.copy-all-button .icon {
  margin-right: 5px;
}
.shotgun-button {
  margin-top: 10px;
  padding: 8px 15px;
  cursor: pointer;
}
.project-path {
  margin: 10px 0;
  font-size: 0.9em;
  color: #555;
}
.error-message {
  color: red;
  margin: 10px 0;
}
</style>
