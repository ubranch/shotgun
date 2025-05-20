<template>
  <div class="p-6 flex flex-col h-full">
    <h2 class="text-xl font-semibold text-gray-800 mb-2">Step 3: Execute Prompt & Split Diff</h2>
    <p class="text-gray-600 mb-2">
      For now, please go to an external LLM provider like Google AI Studio or an equivalent.
      Copy the full project context generated in Step 1 and the prompt you composed in Step 2.
      Paste them into the LLM and obtain the resulting diff output.
    </p>
    <p class="text-gray-600 mb-2">
      Then, paste the full <code>shotgunDiff</code> XML (the LLM's response) below.
      You can then specify the approximate number of lines per split, or leave it as the total number of lines if you don't want to split the diff.
    </p>

    <p class="text-gray-600 mb-2">
      <strong>Why Split the Diff?</strong>
      <br>
      Sometimes, <code>shotgunDiff</code> is a large file that can be difficult to apply with some LLMs. Splitting it into smaller parts makes it easier to apply and reduces the risk of errors.
    </p>

    <hr class="my-4"/>

    <div class="mb-4">
      <label for="shotgun-diff-input" class="block text-sm font-bold text-gray-700 mb-1">Shotgun Diff XML:</label>
      <textarea
        id="shotgun-diff-input"
        v-model="shotgunDiffInput"
        rows="15"
        class="w-full p-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm font-mono"
        placeholder="<shotgunDiff xmlns=...>"
        @input="highlightLines"
      ></textarea>
    </div>

    <div class="mb-4">
      <label for="split-line-limit" class="block text-sm font-bold text-gray-700 mb-1">Approx. Lines per Split:</label>
      <p class="text-gray-600 mb-2 text-xs">
        ⓘ This will attempt to split the diff into the specified number of lines, while keeping the original structure and the hunks.
        The exact number of lines per split is not guaranteed, but the diff will be split into as many parts as possible.
        <br>
        Leave this unchanged if you don't want to split the diff.
      </p>
      <input
        type="number"
        id="split-line-limit"
        v-model.number="splitLineLimit"
        min="50"
        step="50"
        class="w-1/8 p-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm"
      />
      <p class="text-gray-600 mb-2 text-xs mt-2">
        Total number of lines: {{ shotgunDiffInputLines }} <a href="#" class="text-blue-500" title="Reset to this value" @click="resetSplitLineLimit">(reset to this value)</a>
      </p>
    </div>

    <button
      @click="handleSplitDiff"
      :disabled="!shotgunDiffInput.trim() || splitLineLimit <= 0"
      class="px-6 py-2 bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 self-start disabled:bg-gray-400"
    >
      {{ splitLineLimit === shotgunDiffInputLines ? 'Proceed to Apply' : 'Split Diff & Proceed to Apply' }}
    </button>
  </div>
</template>

<style scoped>
  .highlight { color: green; }
</style>

<script setup>
import { ref, defineEmits, watch } from 'vue';

const emit = defineEmits(['action']);

const shotgunDiffInput = ref('');
const splitLineLimit = ref(500); // Initial default
const shotgunDiffInputLines = ref(0);

watch(shotgunDiffInput, (newVal) => {
  if (newVal && newVal.trim() !== '') {
    const lines = newVal.split('\n').length;
    shotgunDiffInputLines.value = lines;
    splitLineLimit.value = lines > 0 ? lines : 500; // Set to line count, or 500 if empty/no lines
  } else {
    // Reset to default if input is cleared
    shotgunDiffInputLines.value = 0;
    splitLineLimit.value = 500; 
  }
  highlightLines();
}, { immediate: false }); // 'immediate: false' to avoid running on initial setup with empty input unless desired

function handleSplitDiff() {
  if (!shotgunDiffInput.value.trim() || splitLineLimit.value <= 0) {
    // Basic validation, could add more specific error messages to user
    return;
  }
  emit('action', 'splitDiff', {
    diffXML: shotgunDiffInput.value,
    lineLimit: splitLineLimit.value
  });
}

const highlightLines = () => {
  const textarea = document.getElementById('shotgun-diff-input');
  const lines = textarea.value.split('\n');
  const highlightedLines = lines.map(line => line.startsWith('+') ? `<span class="highlight">${line}</span>` : line).join('\n');
  textarea.innerHTML = highlightedLines;
};

const resetSplitLineLimit = () => {
  splitLineLimit.value = shotgunDiffInputLines.value;
}

// Expose for MainLayout to potentially clear/set these values if needed
defineExpose({ shotgunDiffInput, splitLineLimit });
</script> 