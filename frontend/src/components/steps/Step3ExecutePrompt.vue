<template>
  <div class="p-6 flex flex-col h-full">
    <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-2">step 3: execute prompt</h2>
    <p class="text-gray-600 dark:text-gray-400 mb-2">
      <li>for now go to google ai studio, copy the prompt and paste it there with 2.5 pro model with 0.1 temperature. it will give you <b>the diff</b></li>
      <li>then open any agentic code tool and ask 'apply diff' + copy-paste the diff. </li>
    </p>
    <p class="text-gray-600 dark:text-gray-400 mb-2">
    <hr class="my-4 border-gray-300 dark:border-gray-700"/>
      <strong>prepare the diff to apply</strong>
      <br>
      this tool will split the diff into smaller parts to make it easier to apply.
    </p>
    <div class="mb-4">
      <label for="shotgun-git-diff-input" class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-1">git diff output:</label>
      <textarea
        id="shotgun-git-diff-input"
        v-model="localShotgunGitDiffInput"
        rows="15"
        class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm font-mono bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        placeholder="paste the git diff output here, e.g., diff --git a/file.txt b/file.txt..."
      ></textarea>
    </div>

    <div class="mb-4">
      <label for="split-line-limit" class="block text-sm font-bold text-gray-700 dark:text-gray-300 mb-1">approx. lines per split:</label>
      <p class="text-gray-600 dark:text-gray-400 mb-2 text-xs">
        ⓘ this will attempt to split the diff into the specified number of lines, while keeping the original structure and the hunks.
        the exact number of lines per split is not guaranteed, but the diff will be split into as many parts as possible.
        <br>
        leave this unchanged if you don't want to split the diff.
      </p>
      <input
        type="number"
        id="split-line-limit"
        v-model.number="localSplitLineLimit"
        min="50"
        step="50"
        class="w-1/8 p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
      />
      <p class="text-gray-600 dark:text-gray-400 mb-2 text-xs mt-2">
        total number of lines: {{ shotgunGitDiffInputLines }} <a href="#" class="text-blue-500 dark:text-blue-400" title="reset to this value" @click="resetSplitLineLimit">(reset to this value)</a>
      </p>
    </div>

    <button
      @click="handleSplitDiff"
      :disabled="!localShotgunGitDiffInput.trim() || localSplitLineLimit <= 0"
      class="px-6 py-2 bg-blue-600 dark:bg-blue-700 text-white font-semibold rounded-md hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 self-start disabled:bg-gray-400 dark:disabled:bg-gray-700"
    >
      {{ localSplitLineLimit === shotgunGitDiffInputLines ? 'proceed to apply' : 'split diff & proceed to apply' }}
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
    // clear any pending debounced updates
  clearTimeout(diffInputDebounceTimer);
  clearTimeout(limitDebounceTimer);

  // immediately emit the current value of localshotgungitdiffinput if it's different from the prop
    if (localShotgunGitDiffInput.value !== props.initialGitDiff) {
        emit('update:shotgunGitDiff', localShotgunGitDiffInput.value);
  } else {
       }

  // immediately emit the current value of localsplitlinelimit if it's valid and different from the prop
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
