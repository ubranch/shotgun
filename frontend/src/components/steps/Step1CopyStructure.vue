<template>
  <div class="p-4 h-full flex flex-col">
    <button
      @click="$emit('action', 'prepareContext')"
      class="px-6 py-2 bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 mb-4 self-start"
      :disabled="!projectRoot"
      :class="{ 'opacity-50 cursor-not-allowed': !projectRoot }"
    >
      Prepare Project Context & Proceed
    </button>

    <div v-if="step1ContextGenerationAttempted" class="mt-2 flex-grow flex flex-col">
      <h3 v-if="generatedContext" class="text-md font-medium text-gray-700 mb-2">Generated Project Context:</h3>
      <textarea
        :value="generatedContext"
        rows="10"
        readonly
        class="w-full p-2 border border-gray-300 rounded-md shadow-sm bg-gray-50 font-mono text-xs flex-grow"
        placeholder="Context will appear here after generation. Ensure files are selected in the sidebar if it's empty."
        style="min-height: 150px;"
      ></textarea>
      <button
        v-if="generatedContext"
        @click="copyGeneratedContextToClipboard"
        class="mt-2 px-4 py-1 bg-gray-200 text-gray-700 font-semibold rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 self-start"
      >
        {{ copyButtonText }}
      </button>
      <p v-if="!generatedContext && step1ContextGenerationAttempted" class="text-xs text-gray-500 mt-2">
        No context generated. Ensure files/folders are selected in the sidebar and are not all excluded, then try again.
      </p>
    </div>
  </div>
</template>

<script setup>
import { defineEmits, defineProps, ref } from 'vue';

defineEmits(['action']);

const props = defineProps({
  generatedContext: {
    type: String,
    default: ''
  },
  step1ContextGenerationAttempted: {
    type: Boolean,
    default: false
  },
  projectRoot: {
    type: String,
    default: ''
  }
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