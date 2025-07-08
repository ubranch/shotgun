<template>
    <div class="p-4 h-full flex flex-col">
        <!-- loading state: always progress bar -->
        <div
            v-if="isLoadingContext"
            class="flex-grow flex justify-center items-center"
        >
            <div class="text-center">
                <div class="w-64 mx-auto">
                    <p class="text-gray-600 dark:text-gray-400 mb-1 text-sm">
                        generating project context...
                    </p>
                    <div
                        class="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700"
                    >
                        <div
                            class="bg-light-accent dark:bg-dark-accent h-2.5 rounded-full"
                            :style="{ width: progressBarWidth }"
                        ></div>
                    </div>
                    <p class="text-gray-500 dark:text-gray-400 mt-1 text-sm">
                        {{ generationProgress.current }} /
                        {{
                            generationProgress.total > 0
                                ? generationProgress.total
                                : "calculating..."
                        }}
                        items
                    </p>
                </div>
            </div>
        </div>

        <!-- content area (textarea + copy button or error message or placeholder) -->
        <div v-else-if="projectRoot" class="mt-0 flex-grow flex flex-col">
            <div
                v-if="isErrorContext"
                class="flex-grow flex flex-col justify-center items-center"
            >
                <div class="max-w-3xl w-full p-6 border border-red-300 dark:border-red-700 rounded-lg bg-red-50 dark:bg-red-900 dark:bg-opacity-20 shadow-sm">
                    <h4 class="text-lg font-semibold mb-3 text-red-600 dark:text-red-400">error generating context</h4>
                    <pre
                        class="text-sm whitespace-pre-wrap text-left w-full bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 p-4 border border-red-200 dark:border-red-700 rounded-md overflow-auto max-h-[50vh]"
                    >{{ errorMessage }}</pre>
                    <p class="mt-4 text-sm text-gray-600 dark:text-gray-400">try reducing the project scope by excluding more files or using a smaller project</p>
                </div>
            </div>
            <div
                v-else-if="generatedContext && !isErrorContext"
                class="flex-grow flex flex-col"
            >
                <div class="flex justify-between items-center mb-2">
                    <h3
                        class="text-md font-medium text-gray-700 dark:text-gray-300"
                    >
                        generated project context:
                    </h3>
                    <button
                        v-if="generatedContext"
                        @click="copyGeneratedContextToClipboard"
                        class="px-3 py-2 bg-light-accent dark:bg-dark-accent text-white text-sm font-semibold rounded-md hover:bg-light-accent-hover dark:hover:bg-dark-accent-hover focus:outline-none disabled:bg-gray-300 dark:disabled:bg-gray-700 flex items-center gap-1"
                        :class="{'bg-green-600 dark:bg-green-700': copySuccess}"
                    >
                        <svg v-if="!copySuccess" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                        </svg>
                        <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                        </svg>
                        {{ copyButtonText }}
                    </button>
                </div>
                <textarea
                    :value="generatedContext"
                    rows="10"
                    readonly
                    class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-gray-50 dark:bg-dark-surface font-mono text-sm text-gray-900 dark:text-gray-100 flex-grow"
                    placeholder="context will appear here. if empty, ensure files are selected and not all excluded."
                    style="min-height: 150px"
                ></textarea>
            </div>
            <p
                v-else
                class="text-sm text-gray-500 dark:text-gray-400 mt-2 flex-grow flex justify-center items-center"
            >
                project context will be generated automatically. if empty after
                generation, ensure files are selected and not all excluded.
            </p>
        </div>

        <!-- initial message when no project is selected -->
        <p
            v-else
            class="text-sm text-gray-500 dark:text-gray-400 mt-2 flex-grow flex justify-center items-center"
        >
            select a project folder to begin.
        </p>
    </div>
</template>

<script setup>
import { defineProps, ref, computed } from "vue";
import { ClipboardSetText as WailsClipboardSetText } from "../../../wailsjs/runtime/runtime";

const props = defineProps({
    generatedContext: {
        type: String,
        default: "",
    },
    projectRoot: {
        type: String,
        default: "",
    },
    isLoadingContext: {
        // new prop
        type: Boolean,
        default: false,
    },
    generationProgress: {
        // new prop for progress data
        type: Object,
        default: () => ({ current: 0, total: 0 }),
    },
    platform: {
        // to know if we are on macos
        type: String,
        default: "unknown",
    },
});

const isErrorContext = computed(() => {
    if (!props.generatedContext) return false;
    // consider only the first non-blank line to decide if the backend sent an error
    const firstLine = props.generatedContext.trimStart().split('\n', 1)[0].toLowerCase();
    return firstLine.startsWith('error:');
});

const errorMessage = computed(() => {
    if (!isErrorContext.value || !props.generatedContext) return '';

    // check if starts with "Error:" (case insensitive)
    const lowerCaseContext = props.generatedContext.toLowerCase();
    if (lowerCaseContext.startsWith('error:')) {
        return props.generatedContext.substring(props.generatedContext.indexOf(':') + 1).trim();
    }

    // if it contains "error:" elsewhere, try to extract the message
    if (lowerCaseContext.includes('error:')) {
        const errorIndex = lowerCaseContext.indexOf('error:');
        return props.generatedContext.substring(errorIndex).trim();
    }

    return props.generatedContext.trim();
});

const progressBarWidth = computed(() => {
    if (props.generationProgress && props.generationProgress.total > 0) {
        const percentage =
            (props.generationProgress.current /
                props.generationProgress.total) *
            100;
        return `${Math.min(100, Math.max(0, percentage))}%`;
    }
    return "0%";
});
const copyButtonText = ref("copy");
const copySuccess = ref(false);

async function copyGeneratedContextToClipboard() {
    if (!props.generatedContext) return;
    try {
        await navigator.clipboard.writeText(props.generatedContext);
        //if (props.platform === 'darwin') {
        //  await WailsClipboardSetText(props.generatedContext);
        //} else {
        //  await navigator.clipboard.writeText(props.generatedContext);
        //}
        copyButtonText.value = "copied!";
        copySuccess.value = true;
        setTimeout(() => {
            copyButtonText.value = "copy";
            copySuccess.value = false;
        }, 2000);
    } catch (err) {
        console.error("failed to copy context: ", err);
        if (props.platform === "darwin" && err) {
            console.error("darvin clipboardsettext failed for context:", err);
        }
        copyButtonText.value = "failed!";
        copySuccess.value = false;
        setTimeout(() => {
            copyButtonText.value = "copy";
        }, 2000);
    }
}
</script>
