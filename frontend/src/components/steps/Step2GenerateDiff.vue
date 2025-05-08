<template>
  <div class="p-6 flex flex-col h-full">
    <h2 class="text-xl font-semibold text-gray-800 mb-4">Step 2: Compose Prompt</h2>
    <p class="text-gray-600 mb-2">
      Provide your prompt for the LLM. Based on this and the "copied" structure, a diff will be generated.
    </p>
    
    <div class="mb-4 border border-gray-300 rounded-md p-3 bg-gray-50 max-h-64 overflow-auto">
      <div class="flex justify-between items-center mb-1">
        <h3 class="text-md font-medium text-gray-700">Project Context (Shotgun Output):</h3>
        <button @click="copyContextToClipboard" class="text-xs px-2 py-1 bg-gray-200 hover:bg-gray-300 rounded">
          {{ copyButtonText }}
        </button>
      </div>
      <pre v-if="shotgunContext" class="text-xs whitespace-pre-wrap font-mono bg-white p-2 rounded border border-gray-200">{{ shotgunContext }}</pre>
      <p v-else class="text-xs text-gray-400 italic">Project context will appear here after Step 1.</p>
    </div>

    <div class="mb-4 flex-grow flex flex-col">
      <label for="prompt-editor" class="block text-sm font-medium text-gray-700 mb-1">Prompt Editor:</label>
      <textarea
        id="prompt-editor"
        v-model="promptText"
        rows="5"
        class="w-full p-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
        placeholder="Enter your instructions for the LLM..."
      ></textarea>
    </div>

    <button
      @click="$emit('action', 'composePrompt', { prompt: promptText })"
      class="px-6 py-2 bg-indigo-600 text-white font-semibold rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50 self-start mb-4"
    >
      Compose Prompt
    </button>

    <div class="border border-gray-300 rounded-md p-4 bg-gray-50 overflow-auto min-h-[100px]">
      <h3 class="text-lg font-medium text-gray-700 mb-2">Diff Viewer:</h3>
      <pre v-if="diffOutput" class="text-sm whitespace-pre-wrap font-mono">{{ diffOutput }}</pre>
      <p v-else class="text-sm text-gray-500">Diff output will appear here after generation.</p>
    </div>
  </div>
</template>

<script setup>
import { ref, defineEmits } from 'vue';

const promptText = ref('');
const diffOutput = ref(''); 
const shotgunContext = ref('');
const copyButtonText = ref('Copy Context');

defineEmits(['action']);

const setDiffOutput = (output) => { diffOutput.value = output; };
const setShotgunContext = (context) => { shotgunContext.value = context; };

async function copyContextToClipboard() {
  if (!shotgunContext.value) return;
  try {
    await navigator.clipboard.writeText(shotgunContext.value);
    copyButtonText.value = 'Copied!';
    setTimeout(() => { copyButtonText.value = 'Copy Context'; }, 2000);
  } catch (err) {
    console.error('Failed to copy context: ', err);
    copyButtonText.value = 'Failed!';
    setTimeout(() => { copyButtonText.value = 'Copy Context'; }, 2000);
  }
}

defineExpose({ setDiffOutput, setShotgunContext });
</script> 