<template>
  <div class="p-4 h-full flex flex-col">
    <p class="text-gray-700 mb-4 text-center text-sm">
      Write the task for the LLM in the central column and copy the final prompt
    </p>

    <CustomRulesModal
      :is-visible="isPromptRulesModalVisible"
      :initial-rules="currentPromptRulesForModal"
      title="Edit Custom Prompt Rules"
      ruleType="prompt"
      @save="handleSavePromptRules"
      @cancel="handleCancelPromptRules"
    />

    <div class="flex-grow flex flex-row space-x-4 overflow-hidden">
      <div class="w-1/2 flex flex-col space-y-3 overflow-y-auto p-2 border border-gray-200 rounded-md bg-gray-50">
        <div>
          <label for="user-task-ai" class="block text-sm font-medium text-gray-700 mb-1">Your task for AI:</label>
          <textarea
            id="user-task-ai"
            v-model="userTask"
            rows="15"
            class="w-full p-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm"
            placeholder="Describe what the AI should do..."
          ></textarea>
        </div>

        <div>
          <label for="rules-content" class="block text-sm font-medium text-gray-700 mb-1 flex items-center">
            Custom rules:
            <button @click="openPromptRulesModal" title="Edit custom prompt rules" class="ml-2 p-0.5 hover:bg-gray-200 rounded text-xs">⚙️</button>
          </label>
          <textarea
            id="rules-content"
            v-model="rulesContent"
            rows="8"
            class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-100 text-sm font-mono"
            placeholder="Rules for AI..."
          ></textarea>
        </div>

        <div>
          <label for="file-list-context" class="block text-sm font-medium text-gray-700 mb-1">Files to include:</label>
          <textarea
            id="file-list-context"
            :value="props.fileListContext"
            rows="20"
            readonly
            class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-100 font-mono text-xs"
            placeholder="File list from Step 1 (Prepare Context) will appear here..."
            style="min-height: 150px;"
          ></textarea>
        </div>
      </div>

      <div class="w-1/2 flex flex-col overflow-y-auto p-2 border border-gray-200 rounded-md bg-white">
        <div class="flex justify-between items-center mb-2">
          <h3 class="text-md font-medium text-gray-700">Final Prompt:</h3>
          <div class="flex items-center space-x-3">
            <span
              v-show="!isLoadingFinalPrompt"
              :class="['text-xs font-medium', charCountColorClass]"
              :title="tooltipText"
            >
              {{ formattedCharCount }}
            </span>
            <button
              @click="copyFinalPromptToClipboard"
              :disabled="!finalPrompt || isLoadingFinalPrompt"
              class="px-3 py-1 bg-blue-500 text-white text-xs font-semibold rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 disabled:bg-gray-300"
            >
              {{ copyButtonText }}
            </button>
          </div>
        </div>
        <div v-if="isLoadingFinalPrompt" class="flex-grow flex justify-center items-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          <p class="text-gray-500 ml-2">Updating prompt...</p>
        </div>
        <textarea
          v-else
          v-model="finalPrompt"
          rows="20"
          class="w-full p-2 border border-gray-300 rounded-md shadow-sm font-mono text-xs flex-grow"
          placeholder="The final prompt will be generated here..."
          style="min-height: 300px;"
        ></textarea>
         <p class="text-xs text-gray-500 mt-1">
            The prompt updates automatically. Manual changes to this field may be overwritten when source data (task, rules, file list) is updated.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';
import { ClipboardSetText as WailsClipboardSetText } from '../../../wailsjs/runtime/runtime';
import { GetCustomPromptRules, SetCustomPromptRules } from '../../../wailsjs/go/main/App';
import { LogInfo as LogInfoRuntime, LogError as LogErrorRuntime } from '../../../wailsjs/runtime/runtime';
import CustomRulesModal from '../CustomRulesModal.vue';
import promptTemplateContentFromFile from '../../../../design/prompts/prompt_makeDiff6.md?raw';

const props = defineProps({
  fileListContext: {
    type: String,
    default: ''
  },
  platform: { // To know if we are on macOS
    type: String,
    default: 'unknown'
  }
});

const emit = defineEmits(['update:finalPrompt']);

const userTask = ref('');
const rulesContent = ref('');
const finalPrompt = ref('');
const isLoadingFinalPrompt = ref(false);
const copyButtonText = ref('Copy All');

let debounceTimer = null;

// Modal state for prompt rules
const isPromptRulesModalVisible = ref(false);
const currentPromptRulesForModal = ref('');

// Character count and related computed properties
const charCount = computed(() => {
  return (finalPrompt.value || '').length;
});

const formattedCharCount = computed(() => {
  return charCount.value.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
});

const charCountColorClass = computed(() => {
  const count = charCount.value;
  if (count < 1000000) {
    return 'text-green-600';
  } else if (count <= 4000000) {
    return 'text-yellow-500'; // Using 500 for better visibility on white bg
  } else {
    return 'text-red-600';
  }
});

const tooltipText = computed(() => {
  if (isLoadingFinalPrompt.value) return 'Calculating...';
  
  const count = charCount.value;
  const tokens = Math.round(count / 3.5);
  return `Your text contains ${count} symbols which is roughly equivalent to ${tokens} tokens`;
});

const PROMPT_TEMPLATE = promptTemplateContentFromFile;

const DEFAULT_RULES = `no additional rules`;

onMounted(async () => {
  try {
    const fetchedRules = await GetCustomPromptRules();
    rulesContent.value = fetchedRules; // Go side ensures a default if empty
  } catch (error) {
    console.error("Failed to load custom prompt rules:", error);
    LogErrorRuntime(`Failed to load custom prompt rules: ${error.message || error}`);
    rulesContent.value = DEFAULT_RULES; // Fallback for actual errors
  }

  if (props.fileListContext || userTask.value) {
    debouncedUpdateFinalPrompt();
  }
});

async function updateFinalPrompt() {
  isLoadingFinalPrompt.value = true;
  await new Promise(resolve => setTimeout(resolve, 100));

  let populatedPrompt = PROMPT_TEMPLATE;
  populatedPrompt = populatedPrompt.replace('{TASK}', userTask.value || "No task provided by the user.");
  populatedPrompt = populatedPrompt.replace('{RULES}', rulesContent.value);
  populatedPrompt = populatedPrompt.replace('{FILE_STRUCTURE}', props.fileListContext || "No file structure context provided.");

  finalPrompt.value = populatedPrompt;
  emit('update:finalPrompt', finalPrompt.value);
  isLoadingFinalPrompt.value = false;
}

function debouncedUpdateFinalPrompt() {
  clearTimeout(debounceTimer);
  isLoadingFinalPrompt.value = true;
  debounceTimer = setTimeout(() => {
    updateFinalPrompt();
  }, 750);
}

watch([userTask, rulesContent, () => props.fileListContext], () => {
  debouncedUpdateFinalPrompt();
}, { deep: true });

async function copyFinalPromptToClipboard() {
  if (!finalPrompt.value) return;
  try {
    await navigator.clipboard.writeText(finalPrompt.value);
    //if (props.platform === 'darwin') {
    //  await WailsClipboardSetText(finalPrompt.value);
    //} else {
    //  await navigator.clipboard.writeText(finalPrompt.value);
    //}
    copyButtonText.value = 'Copied!';
    setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000);
  } catch (err) {
    console.error('Failed to copy final prompt: ', err);
    if (props.platform === 'darwin' && err) {
      console.error('darvin ClipboardSetText failed for final prompt:', err);
    }
    copyButtonText.value = 'Failed!';
    setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000);
  }
}

async function openPromptRulesModal() {
  try {
    currentPromptRulesForModal.value = await GetCustomPromptRules();
    isPromptRulesModalVisible.value = true;
  } catch (error) {
    console.error("Error fetching prompt rules for modal:", error);
    LogErrorRuntime(`Error fetching prompt rules for modal: ${error.message || error}`);
    // Fallback to current editor content or a default if rulesContent is also problematic
    currentPromptRulesForModal.value = rulesContent.value || DEFAULT_RULES;
    isPromptRulesModalVisible.value = true; // Still open modal but with potentially stale/default data
  }
}

async function handleSavePromptRules(newRules) {
  try {
    await SetCustomPromptRules(newRules);
    rulesContent.value = newRules;
    isPromptRulesModalVisible.value = false;
    LogInfoRuntime('Custom prompt rules saved successfully.');
    // The watcher on rulesContent will trigger debouncedUpdateFinalPrompt
  } catch (error) {
    console.error("Error saving prompt rules:", error);
    LogErrorRuntime(`Error saving prompt rules: ${error.message || error}`);
    // Optionally, keep modal open or show an error message to the user
  }
}

function handleCancelPromptRules() {
  isPromptRulesModalVisible.value = false;
}

defineExpose({});
</script>