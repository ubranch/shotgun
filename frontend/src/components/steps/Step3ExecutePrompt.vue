<template>
  <div class="p-6 flex flex-col h-full">
    <h2 class="text-xl font-semibold text-gray-800 mb-2">Step 3: Execute Prompt & Split Diff</h2>
    <p class="text-gray-600 mb-2">
      For now, please go to an external LLM provider like Google AI Studio or an equivalent.
      Copy the full project context generated in Step 1 and the prompt you composed in Step 2.
      Paste them into the LLM and obtain the resulting diff output.
    </p>
    <p class="text-gray-600 mb-4">
      Then, paste the full <code>gitDiff</code> output (the LLM's response) below.
      You can also specify the approximate number of lines per split, or leave it as the total number of lines if you don't want to split the diff.
    </p>

    <p class="text-gray-600 mb-2">
      <strong>Why Split the Diff?</strong>
      <br>
      Sometimes, the generated diff is a large file that can be difficult to apply with some LLMs or review tools. Splitting it into smaller parts makes it easier to manage and reduces the risk of errors.
    </p>

    <hr class="my-4"/>

    <div class="mb-4">
      <label for="shotgun-git-diff-input" class="block text-sm font-bold text-gray-700 mb-1">Git Diff Output:</label>
      <textarea
        id="shotgun-git-diff-input"
        v-model="localShotgunGitDiffInput"
        rows="15"
        class="w-full p-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm font-mono"
        placeholder="Paste the git diff output here, e.g., diff --git a/file.txt b/file.txt..."
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
        v-model.number="localSplitLineLimit"
        min="50"
        step="50"
        class="w-1/8 p-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm"
      />
      <p class="text-gray-600 mb-2 text-xs mt-2">
        Total number of lines: {{ shotgunGitDiffInputLines }} <a href="#" class="text-blue-500" title="Reset to this value" @click="resetSplitLineLimit">(reset to this value)</a>
      </p>
    </div>

    <button
      @click="handleSplitDiff"
      :disabled="!localShotgunGitDiffInput.trim() || localSplitLineLimit <= 0"
      class="px-6 py-2 bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 self-start disabled:bg-gray-400"
    >
      {{ localSplitLineLimit === shotgunGitDiffInputLines ? 'Proceed to Apply' : 'Split Diff & Proceed to Apply' }}
    </button>
  </div>
</template>

<script setup>
import { ref, defineEmits, watch, computed, onMounted, onBeforeUnmount } from 'vue';
import { LogInfo as LogInfoRuntime, LogError as LogErrorRuntime } from '../../../wailsjs/runtime/runtime';

const emit = defineEmits(['action', 'update:shotgunGitDiff', 'update:splitLineLimit']);

const props = defineProps({
  initialGitDiff: {
    type: String,
    default: ''
  },
  initialSplitLineLimit: {
    type: Number,
    default: 0
  }
});


const localShotgunGitDiffInput = ref(props.initialGitDiff);

const localSplitLineLimit = ref(props.initialSplitLineLimit > 0 ? props.initialSplitLineLimit : 500);

onMounted(() => {
    
  localShotgunGitDiffInput.value = props.initialGitDiff;

    
  if (props.initialSplitLineLimit > 0) {
    localSplitLineLimit.value = props.initialSplitLineLimit;
  } else if (localSplitLineLimit.value <= 0) {
    localSplitLineLimit.value = 500;
  }
});

const shotgunGitDiffInputLines = computed(() => {
  return localShotgunGitDiffInput.value ? localShotgunGitDiffInput.value.split('\n').length : 0;
});

watch(() => props.initialGitDiff, (newVal, oldVal) => {
        if (newVal !== localShotgunGitDiffInput.value) {
                localShotgunGitDiffInput.value = newVal;
            }
});

watch(() => props.initialSplitLineLimit, (newVal, oldVal) => {
        if (newVal > 0 && newVal !== localSplitLineLimit.value) {
        localSplitLineLimit.value = newVal;
    } else if (newVal <= 0 && localSplitLineLimit.value !== 500 && props.initialGitDiff === '') {
        localSplitLineLimit.value = 500;
    }
});

let diffInputDebounceTimer = null;
watch(localShotgunGitDiffInput, (newVal, oldVal) => {
    
    clearTimeout(diffInputDebounceTimer);
    
    diffInputDebounceTimer = setTimeout(() => {
                if (newVal !== props.initialGitDiff) {
                        emit('update:shotgunGitDiff', newVal);
        } else {
                    }
        if (newVal && newVal.trim() !== '') {
            const lines = newVal.split('\n').length;
            const currentLimit = localSplitLineLimit.value;

            if (currentLimit === 500 || (currentLimit !== lines && currentLimit === (newVal.substring(0, newVal.length - (newVal.split('\n').pop().length +1)).split('\n').length))) {
                if (lines > 0 && lines !== currentLimit) {
                    localSplitLineLimit.value = lines;
                }
            } else if (lines === 0 && currentLimit !== 500){
                 localSplitLineLimit.value = 500;
            }
        } else if ((!newVal || newVal.trim() === '') && localSplitLineLimit.value !== 500) {
            localSplitLineLimit.value = 500;
        }
    }, 300);
});

let limitDebounceTimer = null;
watch(localSplitLineLimit, (newVal) => {
    clearTimeout(limitDebounceTimer);
    limitDebounceTimer = setTimeout(() => {
        if (newVal > 0 && newVal !== props.initialSplitLineLimit) { 
            emit('update:splitLineLimit', newVal);
        } else if (newVal <= 0 && props.initialSplitLineLimit > 0) {
        }
    }, 300);
});

onBeforeUnmount(() => {
    // Clear any pending debounced updates
  clearTimeout(diffInputDebounceTimer);
  clearTimeout(limitDebounceTimer);
  
  // Immediately emit the current value of localShotgunGitDiffInput if it's different from the prop
    if (localShotgunGitDiffInput.value !== props.initialGitDiff) {
        emit('update:shotgunGitDiff', localShotgunGitDiffInput.value);
  } else {
       }

  // Immediately emit the current value of localSplitLineLimit if it's valid and different from the prop
    if (localSplitLineLimit.value > 0 && localSplitLineLimit.value !== props.initialSplitLineLimit) {
        emit('update:splitLineLimit', localSplitLineLimit.value);
  } else {
      }
});

function handleSplitDiff() {
  if (!localShotgunGitDiffInput.value.trim() || localSplitLineLimit.value <= 0) {
    return;
  }
  emit('action', 'executePromptAndSplitDiff', {
    gitDiff: localShotgunGitDiffInput.value,
    lineLimit: localSplitLineLimit.value
  });
}

const resetSplitLineLimit = () => {
  if (shotgunGitDiffInputLines.value > 0) {
    localSplitLineLimit.value = shotgunGitDiffInputLines.value;
  } else {
    localSplitLineLimit.value = 500;
  }
}
</script> 