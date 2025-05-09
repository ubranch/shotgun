<template>
  <div class="p-4 h-full flex flex-col">
    <!-- Spinner -->
    <div v-if="isLoadingContext" class="flex-grow flex justify-center items-center">
      <div class="text-center">
        <svg class="animate-spin h-8 w-8 text-blue-600 mx-auto mb-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="text-gray-600">Generating project context...</p>
      </div>
    </div>

    <!-- Content Area (Textarea + Copy Button OR Error Message OR Placeholder) -->
    <div v-else-if="projectRoot" class="mt-0 flex-grow flex flex-col">
      <div v-if="generatedContext && !generatedContext.startsWith('Error:')" class="flex-grow flex flex-col">
        <h3 class="text-md font-medium text-gray-700 mb-2">Generated Project Context:</h3>
        <textarea
          :value="generatedContext"
          rows="10"
          readonly
          class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 font-mono text-xs flex-grow"
          placeholder="Context will appear here. If empty, ensure files are selected and not all excluded."
          style="min-height: 150px;"
        ></textarea>
        <button
          v-if="generatedContext"
          @click="copyGeneratedContextToClipboard"
          class="mt-2 px-4 py-1 bg-gray-200 text-gray-700 font-semibold rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 self-start"
        >
          {{ copyButtonText }}
        </button>
      </div>
      <div v-else-if="generatedContext && generatedContext.startsWith('Error:')" class="text-red-500 p-3 border border-red-300 rounded bg-red-50 flex-grow flex flex-col justify-center items-center">
        <h4 class="font-semibold mb-1">Error Generating Context:</h4>
        <pre class="text-xs whitespace-pre-wrap text-left w-full bg-white p-2 border border-red-200 rounded max-h-60 overflow-auto">{{ generatedContext.substring(6).trim() }}</pre>
      </div>
      <p v-else class="text-xs text-gray-500 mt-2 flex-grow flex justify-center items-center">
        Project context will be generated automatically. If empty after generation, ensure files are selected and not all excluded.
      </p>
    </div>

    <!-- Initial message when no project is selected -->
    <p v-else class="text-xs text-gray-500 mt-2 flex-grow flex justify-center items-center">
      Select a project folder to begin.
    </p>
  </div>
</template>

<script setup>
import { defineProps, ref } from 'vue';

const props = defineProps({
  generatedContext: {
    type: String,
    default: ''
  },
  projectRoot: {
    type: String,
    default: ''
  },
  isLoadingContext: { // New prop
    type: Boolean,
    default: false
  },
});

const copyButtonText = ref('Copy All');

async function copyGeneratedContextToClipboard() {
  if (!props.generatedContext) return;
  try {
    await navigator.clipboard.writeText(props.generatedContext);
    copyButtonText.value = 'Copied!';
    setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000);
  } catch (err) {
    console.error('Failed to copy context: ', err);
    copyButtonText.value = 'Failed!';
    setTimeout(() => {
      copyButtonText.value = 'Copy All';
    }, 2000);
  }
}
</script>