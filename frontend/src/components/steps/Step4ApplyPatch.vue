<template>
  <div class="p-6 flex flex-col h-full">
    <h2 class="text-xl font-semibold text-gray-800 mb-4">Step 4: Apply Patch</h2>
    
    <div v-if="isLoading" class="flex-grow flex justify-center items-center">
      <p class="text-gray-600">Loading split diffs...</p>
    </div>
    
    <div v-else-if="splitDiffs && splitDiffs.length > 0" class="flex-grow overflow-y-auto space-y-6">
      <p class="text-gray-600 mb-2">
        The original diff has been split into {{ splitDiffs.length }} smaller diffs.
        Copy each part and apply it using your preferred tool.
      </p>
      <div v-for="(diff, index) in splitDiffs" :key="index" :class="['border border-gray-300 rounded-md p-4', isCopied[index] ? 'bg-green-50' : 'bg-gray-50', 'shadow-sm']">
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-medium text-gray-700">Split {{ index + 1 }} of {{ splitDiffs.length }}</h3>
          <button
            @click="copyDiffToClipboard(diff, index)"
            class="px-3 py-1 bg-gray-200 text-gray-700 text-xs font-semibold rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400"
          >
            {{ copyButtonTexts[index] || 'Copy' }}
          </button>
        </div>
        <p class="text-gray-600 text-xs mb-2">
           <!-- the lines metric will be orange if it's greater than 300, soon to update to get line size from previous step -->
          <div class="inline-block px-2 py-1 rounded-full text-xs" :class="diff.split('\n').length > 300 ? 'bg-orange-100' : 'bg-blue-100'">
            {{ diff.split('\n').length }} lines
          </div>
          <div class="inline-block px-2 py-1 bg-blue-100 rounded-full text-xs ml-2">
            {{ diff.split('\n').filter(line => line.trim().startsWith('<file')).length }} file{{ diff.split('\n').filter(line => line.trim().startsWith('<file')).length === 1 ? '' : 's' }}
          </div>
          <div class="inline-block px-2 py-1 bg-blue-100 rounded-full text-xs ml-2">
            {{ diff.split('\n').filter(line => line.trim().startsWith('<hunk')).length }} hunk{{ diff.split('\n').filter(line => line.trim().startsWith('<hunk')).length === 1 ? '' : 's' }}
          </div>
        </p>
        <textarea
          :value="diff"
          rows="10"
          readonly
          class="w-full p-2 border border-gray-200 rounded-md bg-white font-mono text-xs"
          style="min-height: 150px;"
        ></textarea>
      </div>
    </div>
    
    <div v-else class="flex-grow flex justify-center items-center">
      <p class="text-gray-500">No split diffs to display. Go to Step 3 to split a diff.</p>
    </div>

    <div class="mt-4 pt-4 border-t border-gray-200">
      <h3 class="text-lg font-medium text-gray-700 mb-2">Apply Patch automatically <sup class="text-xs text-white bg-green-500 rounded-md px-1 py-1">COMING SOON</sup></h3>
      <p class="text-gray-600 italic">
        Here you will see an interface to review and apply the main patch, 
        potentially with a diff viewer.
        For now, this section is a placeholder. Click 'Finish' to simulate completing the process.
      </p>
    </div>

    <div class="mt-6 flex space-x-4 flex-shrink-0">
      <button
        @click="$emit('action', 'finishSplitting')"
        class="px-6 py-2 bg-blue-500 text-white font-semibold rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50"
      >
        Finish
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, watch } from 'vue';
// import { ClipboardSetText as WailsClipboardSetText } from '../../../wailsjs/runtime/runtime'; // If needed for specific platforms

const props = defineProps({
  splitDiffs: {
    type: Array,
    default: () => []
  },
  isLoading: { // To indicate if MainLayout is fetching/processing splits
    type: Boolean,
    default: false
  },
  platform: {
    type: String,
    default: 'unknown'
  }
});

defineEmits(['action']);

const copyButtonTexts = ref({});
const isCopied = ref({}); // Tracks if a split has been successfully copied at least once

watch(() => props.splitDiffs, (newVal) => {
  // Reset copy button texts and copied states when diffs change
  const newTexts = {};
  const newCopiedStates = {};
  if (newVal) {
    newVal.forEach((_, index) => {
      newTexts[index] = 'Copy';
      newCopiedStates[index] = false; // Initialize as not copied
    });
  }
  copyButtonTexts.value = newTexts;
  isCopied.value = newCopiedStates;
}, { immediate: true, deep: true }); // Use deep: true if splitDiffs could be mutated internally, though usually props are replaced.


async function copyDiffToClipboard(diffContent, index) {
  if (!diffContent) return;
  try {
    await navigator.clipboard.writeText(diffContent);
    
    isCopied.value[index] = true; // Mark as successfully copied
    copyButtonTexts.value[index] = 'Copied! ✅';

    setTimeout(() => {
      copyButtonTexts.value[index] = 'Copy ✅'; // Persistent "copied" state text
    }, 2000);
  } catch (err) {
    console.error(`Failed to copy diff split ${index + 1}: `, err);
    
    // Temporarily show "Failed!"
    const originalText = isCopied.value[index] ? 'Copy ✅' : 'Copy';
    copyButtonTexts.value[index] = 'Failed!';

    setTimeout(() => {
      copyButtonTexts.value[index] = originalText; // Revert to previous state ("Copy" or "Copy ✅")
    }, 2000);
  }
}
</script> 