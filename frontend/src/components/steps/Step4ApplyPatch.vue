<template>
    <div class="p-6 flex flex-col h-full">
        <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">
            step 4: apply patch
        </h2>

        <div
            v-if="isLoading"
            class="flex-grow flex justify-center items-center"
        >
            <p class="text-gray-600 dark:text-gray-400">
                loading split diffs...
            </p>
        </div>

        <div
            v-else-if="splitDiffs && splitDiffs.length > 0"
            class="flex-grow overflow-y-auto space-y-6"
        >
            <p class="text-gray-600 dark:text-gray-400 mb-2 text-sm">
                the original diff has been split into
                {{ splitDiffs.length }} smaller diffs. copy each part and apply
                it using your preferred tool. with an llm, just tell it to
                <strong>apply the diff</strong>.
            </p>
            <div
                v-for="(diff, index) in splitDiffs"
                :key="index"
                :class="[
                    'border border-gray-300 dark:border-gray-700 rounded-md p-4',
                    isCopied[index]
                        ? 'bg-green-50 dark:bg-green-900 dark:bg-opacity-20'
                        : 'bg-gray-50 dark:bg-dark-surface',
                    'shadow-sm',
                ]"
            >
                <div class="flex justify-between items-center">
                    <h3
                        class="text-lg font-medium text-gray-700 dark:text-gray-300"
                    >
                        split {{ index + 1 }} of {{ splitDiffs.length }}
                    </h3>
                    <div class="flex items-center space-x-2">
                        <!-- soon: add a feature to apply the diff automatically -->
                        <!-- <button
                          class="px-3 py-1 bg-gray-100 text-gray-300 text-sm font-semibold rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
                          disabled
                        >
                          apply diff
                        </button> -->
                        <button
                            @click="copyDiffToClipboard(diff, index)"
                            class="px-3 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 text-sm font-semibold rounded-md hover:bg-gray-300 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-400"
                        >
                            {{ copyButtonTexts[index] || "copy" }}
                        </button>
                    </div>
                </div>
                <div class="text-gray-600 dark:text-gray-400 text-sm mb-2">
                    <!-- the lines metric will be orange if it's greater than props.splitlinelimit + 5%, red if it's greater than props.splitlinelimit + 20%, green if it's less than props.splitlinelimit + 5% -->
                    <!-- calculate this in the vue script below, to simplify the code -->
                    <div
                        class="inline-block px-2 py-1 rounded-full text-sm"
                        :class="getLineMetricClass(diff.split('\n').length)"
                    >
                        {{ diff.split("\n").length }} lines
                    </div>
                    <div
                        class="inline-block px-2 py-1 bg-indigo-100 dark:bg-indigo-900 dark:bg-opacity-50 text-gray-900 dark:text-gray-200 rounded-full text-sm ml-2"
                    >
                        {{ (diff.match(/^diff --git/gm) || []).length }} file{{
                            (diff.match(/^diff --git/gm) || []).length === 1
                                ? ""
                                : "s"
                        }}
                    </div>
                    <div
                        class="inline-block px-2 py-1 bg-indigo-100 dark:bg-indigo-900 dark:bg-opacity-50 text-gray-900 dark:text-gray-200 rounded-full text-sm ml-2"
                    >
                        {{ (diff.match(/^@@ .* @@/gm) || []).length }} hunk{{
                            (diff.match(/^@@ .* @@/gm) || []).length === 1
                                ? ""
                                : "s"
                        }}
                    </div>
                </div>
                <textarea
                    :value="diff"
                    rows="10"
                    readonly
                    class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 font-mono text-sm"
                    style="min-height: 150px"
                ></textarea>
            </div>
        </div>

        <div v-else class="flex-grow flex justify-center items-center">
            <p class="text-gray-500 dark:text-gray-400">
                no split diffs to display. go to step 3 to split a diff.
            </p>
        </div>

    </div>
</template>

<script setup>
const finishButtonText = ref("finish");
import { ref, defineProps, watch } from "vue";
// import { clipboardsettext as wailsclipboardsettext } from '../../../wailsjs/runtime/runtime'; // if needed for specific platforms

const props = defineProps({
    splitDiffs: {
        type: Array,
        default: () => [],
    },
    isLoading: {
        // to indicate if mainlayout is fetching/processing splits
        type: Boolean,
        default: false,
    },
    platform: {
        type: String,
        default: "unknown",
    },
    splitLineLimit: {
        // add the new prop
        type: Number,
        default: 500, // provide a default value if the prop is not passed
    },
});

defineEmits(["action"]);

const copyButtonTexts = ref({});
const isCopied = ref({}); // tracks if a split has been successfully copied at least once

function getLineMetricClass(lineCount) {
    const limit = props.splitLineLimit;
    // clamp the thresholds to maximum 100 or 200 lines over the limit
    const orangeThreshold = Math.min(limit * 1.1, limit + 100);
    const redThreshold = Math.min(limit * 1.3, limit + 200);

    if (lineCount > redThreshold) {
        return "bg-red-100 dark:bg-red-900 dark:bg-opacity-30 text-gray-900 dark:text-gray-200";
    } else if (lineCount > orangeThreshold) {
        return "bg-orange-100 dark:bg-orange-900 dark:bg-opacity-30 text-gray-900 dark:text-gray-200";
    } else {
        return "bg-green-100 dark:bg-green-900 dark:bg-opacity-30 text-gray-900 dark:text-gray-200";
    }
}

watch(
    () => props.splitDiffs,
    (newVal) => {
        // reset copy button texts and copied states when diffs change
        const newTexts = {};
        const newCopiedStates = {};
        if (newVal) {
            newVal.forEach((_, index) => {
                newTexts[index] = "copy";
                newCopiedStates[index] = false; // initialize as not copied
            });
        }
        copyButtonTexts.value = newTexts;
        isCopied.value = newCopiedStates;
    },
    { immediate: true, deep: true }
); // use deep: true if splitdiffs could be mutated internally, though usually props are replaced.

async function copyDiffToClipboard(diffContent, index) {
    if (!diffContent) return;
    try {
        await navigator.clipboard.writeText(diffContent);

        isCopied.value[index] = true; // mark as successfully copied
        copyButtonTexts.value[index] = "copied! ✅";

        setTimeout(() => {
            copyButtonTexts.value[index] = "copy ✅"; // persistent "copied" state text
        }, 2000);
    } catch (err) {
        console.error(`failed to copy diff split ${index + 1}: `, err);

        // temporarily show "failed!"
        const originalText = isCopied.value[index] ? "copy ✅" : "copy";
        copyButtonTexts.value[index] = "failed!";

        setTimeout(() => {
            copyButtonTexts.value[index] = originalText; // revert to previous state ("copy" or "copy ✅")
        }, 2000);
    }
}
</script>
