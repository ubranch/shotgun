<template>
  <div class="p-4 h-full flex flex-col">
    <p class="text-gray-700 mb-4 text-center text-sm">
      Write the task for the LLM in the central column and copy the final prompt
    </p>

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
          <label for="rules-content" class="block text-sm font-medium text-gray-700 mb-1">
            Custom rules:
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
  const tokens = Math.round(count / 4);
  return `Your text contains ${count} symbols which is roughly equivalent to ${tokens} tokens`;
});

const PROMPT_TEMPLATE = `## ROLE & PRIMARY GOAL:
You are a "Robotic Senior Software Engineer AI". Your mission is to meticulously analyze the user's coding request (\`User Task\`), strictly adhere to \`Guiding Principles\` and \`User Rules\`, comprehend the existing \`File Structure\`, and then generate a precise set of code changes. Your *sole and exclusive output* must be a single \`<shotgunDiff>\` XML document. Zero tolerance for any deviation from the specified output format.
---
## 1. User Task
{TASK}
---
## 2. Guiding Principles (Your Senior Developer Logic)
### A. Analysis & Planning (Internal Thought Process - Do NOT output this part):
1.  **Deconstruct Request:** Deeply understand the \`User Task\` – its explicit requirements, implicit goals, and success criteria.
2.  **Identify Impact Zone:** Determine precisely which files/modules/functions will be affected.
3.  **Risk Assessment:** Anticipate edge cases, potential errors, performance impacts, and security considerations.
4.  **Assume with Reason:** If ambiguities exist in \`User Task\`, make well-founded assumptions based on best practices and existing code context. Document these assumptions internally if complex.
5.  **Optimal Solution Path:** Briefly evaluate alternative solutions, selecting the one that best balances simplicity, maintainability, readability, and consistency with existing project patterns.
6.  **Plan Changes:** Before generating diffs, mentally (or internally) outline the specific changes needed for each affected file.
### B. Code Generation & Standards:
*   **Simplicity & Idiomatic Code:** Prioritize the simplest, most direct solution. Write code that is idiomatic for the language and aligns with project conventions (inferred from \`File Structure\`). Avoid over-engineering.
*   **Respect Existing Architecture:** Strictly follow the established project structure, naming conventions, and coding style.
*   **Type Safety:** Employ type hints/annotations as appropriate for the language.
*   **Modularity:** Design changes to be modular and reusable where sensible.
*   **Documentation:**
    *   Add concise docstrings/comments for new public APIs, complex logic, or non-obvious decisions.
    *   Update existing documentation if changes render it inaccurate.
*   **Logging:** Introduce logging for critical operations or error states if consistent with the project\'s logging strategy.
*   **No New Dependencies:** Do NOT introduce external libraries/dependencies unless explicitly stated in \`User Task\` or \`User Rules\`.
*   **Atomicity of Hunks:** Each \`<hunk>\` should represent a small, logically coherent change.
*   **Testability:** Design changes to be testable. If a testing framework is evident in \`File Structure\` or mentioned in \`User Rules\`, ensure new code is compatible.
---
## 3. User Rules
{RULES}
---
## 6. File Structure
{FILE_STRUCTURE}
`;

const DEFAULT_RULES = `no additional rules`;

onMounted(() => {
  rulesContent.value = DEFAULT_RULES;
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
    if (props.platform === 'darwin') {
      await WailsClipboardSetText(finalPrompt.value);
    } else {
      await navigator.clipboard.writeText(finalPrompt.value);
    }
    copyButtonText.value = 'Copied!';
    setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000);
  } catch (err) {
    console.error('Failed to copy final prompt: ', err);
    if (props.platform === 'darwin' && err) {
      console.error('Wails ClipboardSetText failed for final prompt:', err);
    }
    copyButtonText.value = 'Failed!';
    setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000);
  }
}

defineExpose({});
</script> 