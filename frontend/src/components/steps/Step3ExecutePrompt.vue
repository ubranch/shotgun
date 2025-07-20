<template>
    <div class="p-6 flex flex-col h-full">
        <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-2">
            step 3: execute & prepare prompt
        </h2>

        <div class="flex flex-row items-center mb-4 space-x-4">
            <BaseButton
                @click="executeRequest"
                class="text-xs px-2 py-1"
                :disabled="isRequestActive || !isReadyToExecute"
            >
                <span class="text-base">execute request</span>
            </BaseButton>
            <BaseButton
                @click="toggleModel"
                class="text-xs px-2 py-1"
                :disabled="isRequestActive || !isReadyToExecute"
                title="click to switch between gemini models"
            >
                <span class="text-base">{{ selectedModel }}</span>
            </BaseButton>
            <!-- token count pill (styled like step 2) -->
            <span v-if="isTokenChecking" class="text-sm text-gray-500 ml-2"
                >counting...</span
            >
            <span
                v-else-if="tokenCountError"
                class="text-sm text-red-500 ml-2"
                :title="tokenCountError"
                >error</span
            >
            <span
                v-else
                :class="[
                    'text-sm font-bold px-2 py-1 rounded-xl',
                    charCountColorClass === 'text-green-600'
                        ? 'bg-green-100 dark:bg-green-900/30'
                        : charCountColorClass === 'text-yellow-500'
                          ? 'bg-yellow-100 dark:bg-yellow-900/30'
                          : 'bg-red-100 dark:bg-red-900/30',
                ]"
                :title="tooltipText"
            >
                {{ promptTokensCount.toLocaleString() }} tokens
            </span>
            <div
                v-if="requestError"
                class="ml-2 max-w-[300px] truncate text-sm rounded-xl bg-red-600 dark:bg-red-700 text-white px-3 py-1 font-mono"
                title="{{ requestError }}"
            >
                {{ requestError }}
            </div>
            <div
                v-if="isPromptTooLarge"
                class="ml-2 max-w-[300px] truncate text-sm rounded-xl bg-red-600 dark:bg-red-700 text-white px-3 py-1 font-mono"
            >
                prompt exceeds free api limit of 250,000 tokens
            </div>
            <div
                v-if="isPromptTooLarge"
                class="ml-2 max-w-[300px] truncate text-sm rounded-xl bg-gray-600 dark:bg-grey-700 text-white px-3 py-1 font-mono"
            >
                use google ai studio instead
            </div>
            <div
                class="flex items-center w-[4rem] justify-center"
                v-if="isRequestActive"
            >
                <div class="text-gray-700 dark:text-gray-300 font-mono">
                    {{ formattedTime }}
                </div>
            </div>
            <div class="w-[4rem] justify-center" v-else></div>
            <BaseButton
                @click="stopRequest"
                class="text-xs px-2 py-1"
                variant="danger"
                v-if="isRequestActive"
            >
                <span class="text-base">stop request</span>
            </BaseButton>
        </div>

        <p class="text-gray-600 dark:text-gray-400 mb-4 text-sm">
            <li>
                open any agentic code tool and ask 'apply diff' + copy-paste the
                diff.
            </li>
            <li>
                use the execute request button above to send the prompt directly
                to gemini api.
            </li>
        </p>

        <hr class="my-4 border-accent" />
        <div class="flex justify-between items-center mb-2">
            <div class="text-gray-600 dark:text-gray-400">
                <strong>prepare the diff to apply</strong>
                <br />
                this tool will split the diff into smaller parts to make it
                easier to apply.
            </div>
            <div class="flex gap-2">
                <BaseButton
                    v-if="localShotgunGitDiffInput.trim()"
                    @click="copyDiffToClipboard"
                    class="text-xs px-2 py-1"
                    :class="{ 'bg-green-600 dark:bg-green-700': copySuccess }"
                >
                    <template #icon>
                        <svg
                            v-if="!copySuccess"
                            xmlns="http://www.w3.org/2000/svg"
                            class="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                            />
                        </svg>
                        <svg
                            v-else
                            xmlns="http://www.w3.org/2000/svg"
                            class="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M5 13l4 4L19 7"
                            />
                        </svg>
                    </template>
                    <span class="text-base">{{ copySuccess ? "copied!" : "copy" }}</span>
                </BaseButton>
                <BaseButton
                    v-if="localShotgunGitDiffInput.trim()"
                    @click="clearTextarea"
                    class="text-xs px-2 py-1"
                    variant="danger"
                    :class="{ 'bg-red-600 dark:bg-red-700': clearSuccess }"
                >
                    <template #icon>
                        <svg
                            v-if="!clearSuccess"
                            xmlns="http://www.w3.org/2000/svg"
                            class="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                            />
                        </svg>
                        <svg
                            v-else
                            xmlns="http://www.w3.org/2000/svg"
                            class="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M5 13l4 4L19 7"
                            />
                        </svg>
                    </template>
                    <span class="text-base">{{ clearSuccess ? "cleared!" : "clear" }}</span>
                </BaseButton>
            </div>
        </div>

        <div class="mb-4">
            <textarea
                id="shotgun-git-diff-input"
                v-model="localShotgunGitDiffInput"
                rows="15"
                spellcheck="false"
                class="w-full p-2 border border-accent rounded-md shadow-sm focus:ring-light-accent dark:focus:ring-dark-accent focus:border-light-accent dark:focus:border-dark-accent text-sm font-mono bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100"
                placeholder="paste the git diff output here, e.g., diff --git a/file.txt b/file.txt..."
            ></textarea>
        </div>

        <div class="mb-4">
            <label
                for="split-line-limit"
                class="block text-base font-bold text-gray-700 dark:text-gray-300 mb-1"
                >approx. lines per split</label
            >
            <p class="text-gray-600 dark:text-gray-400 mb-2 text-sm">
                this will attempt to split the diff into the specified number of
                lines, while keeping the original structure and the chunks.
                <br />
                the exact number of lines per split is not guaranteed, but the
                diff will be split into as many parts as possible.
            </p>
            <div class="flex items-center space-x-2 mt-2">
                <input
                    type="number"
                    id="split-line-limit"
                    v-model.number="localSplitLineLimit"
                    min="50"
                    step="50"
                    class="w-1/8 p-2 border border-accent rounded-md shadow-sm focus:ring-light-accent dark:focus:ring-dark-accent focus:border-light-accent dark:focus:border-dark-accent text-sm bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100"
                />
                <div class="flex items-center gap-2">
                    <span
                        class="px-3 py-1 text-sm font-bold rounded-xl bg-gray-200 dark:bg-gray-600/30 text-gray-700 dark:text-gray-300"
                    >
                        {{ shotgunGitDiffInputLines }} lines in total
                    </span>
                </div>
            </div>
        </div>

        <BaseButton
            @click="handleSplitDiff"
            :disabled="
                !localShotgunGitDiffInput.trim() || localSplitLineLimit <= 0
            "
            class="text-xs px-2 py-1 self-start"
        >
            <span class="text-base">{{
                localSplitLineLimit === shotgunGitDiffInputLines
                    ? "proceed to apply"
                    : "split diff & proceed to apply"
            }}</span>
        </BaseButton>
    </div>
</template>

<script setup>
import {
    ref,
    defineEmits,
    watch,
    computed,
    onMounted,
    onBeforeUnmount,
} from "vue";
import {
    LogInfo as LogInfoRuntime,
    LogError as LogErrorRuntime,
    EventsOn,
    EventsOff,
} from "../../../wailsjs/runtime/runtime";
import {
    ExecuteGeminiRequest,
    StopGeminiRequest,
    CountGeminiTokens,
} from "../../../wailsjs/go/main/App";
import BaseButton from "../BaseButton.vue";

const emit = defineEmits([
    "action",
    "update:shotgunGitDiff",
    "update:splitLineLimit",
]);

const props = defineProps({
    initialGitDiff: {
        type: String,
        default: "",
    },
    initialSplitLineLimit: {
        type: Number,
        default: 0,
    },
    finalPrompt: {
        type: String,
        default: "",
    },
});

// timer state
const isRequestActive = ref(false);
const startTime = ref(null);
const elapsedTime = ref(0);
const timerInterval = ref(null);
const requestError = ref(null);

// copy functionality state
const copySuccess = ref(false);

// clear functionality state
const clearSuccess = ref(false);

// token limit enforcement state
const tokenCountLimit = 250000;
const promptTokensCount = ref(0);
const tokenCountError = ref("");
const isPromptTooLarge = computed(
    () => promptTokensCount.value >= tokenCountLimit
);

// dynamic color class based on token usage (similar to step2)
const charCountColorClass = computed(() => {
    const count = promptTokensCount.value;
    if (count < tokenCountLimit * 0.7) {
        return "text-green-600";
    } else if (count < tokenCountLimit) {
        return "text-yellow-500";
    } else {
        return "text-red-600";
    }
});

// tooltip text similar to step2
const tooltipText = computed(() => {
    if (isTokenChecking.value) return "calculating tokens...";
    if (tokenCountError.value) return `error: ${tokenCountError.value}`;
    return `prompt contains ${promptTokensCount.value.toLocaleString()} gemini tokens`;
});

const isTokenChecking = ref(false);
const isReadyToExecute = computed(() => {
    return !isTokenChecking.value && !isPromptTooLarge.value;
});

// model selection state
const selectedModel = ref("gemini-2.5-pro");

function toggleModel() {
    selectedModel.value = selectedModel.value === "gemini-2.5-pro" 
        ? "gemini-2.5-flash" 
        : "gemini-2.5-pro";
}

const formattedTime = computed(() => {
    const seconds = Math.floor(elapsedTime.value / 1000);
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes.toString().padStart(2, "0")}:${remainingSeconds.toString().padStart(2, "0")}`;
});

async function executeRequest() {
    if (!props.finalPrompt) {
        LogErrorRuntime("no prompt available to execute");
        requestError.value = "no prompt available to execute";
        return;
    }

    isRequestActive.value = true;
    startTime.value = Date.now();
    elapsedTime.value = 0;
    requestError.value = null;

    timerInterval.value = setInterval(() => {
        elapsedTime.value = Date.now() - startTime.value;
    }, 1000);

    try {
        LogInfoRuntime("executing gemini request...");
        // log selected model and request body for transparency
        LogInfoRuntime(`gemini request config: model=${selectedModel.value}`);
        const charCount = props.finalPrompt.length;
        const bodyPreview =
            charCount > 500
                ? props.finalPrompt.slice(0, 500) + "..."
                : props.finalPrompt;
        LogInfoRuntime(`gemini request body length: ${charCount} characters`);
        LogInfoRuntime("gemini request body preview (max 500 chars):");
        LogInfoRuntime(bodyPreview);
        const result = await ExecuteGeminiRequest(
            props.finalPrompt,
            selectedModel.value
        );

        // if we get a result, update the diff input
        if (result) {
            localShotgunGitDiffInput.value = result;
            LogInfoRuntime("gemini request completed successfully");
        }
    } catch (error) {
        // improve error handling to capture both string and error object responses
        const errMsg =
            typeof error === "string"
                ? error
                : error && error.message
                  ? error.message
                  : error?.toString() || "failed to execute gemini request";
        requestError.value = errMsg;
        LogErrorRuntime("gemini request failed: " + errMsg);
    } finally {
        isRequestActive.value = false;
        clearInterval(timerInterval.value);
    }
}

async function stopRequest() {
    if (isRequestActive.value) {
        try {
            await StopGeminiRequest();
            LogInfoRuntime("gemini request stopped by user");
        } catch (error) {
            LogErrorRuntime(
                "failed to stop gemini request: " +
                    (error.message || "unknown error")
            );
        }
    }
}

const localShotgunGitDiffInput = ref(props.initialGitDiff);

const localSplitLineLimit = ref(
    props.initialSplitLineLimit > 0 ? props.initialSplitLineLimit : 500
);

onMounted(() => {
    localShotgunGitDiffInput.value = props.initialGitDiff;

    if (props.initialSplitLineLimit > 0) {
        localSplitLineLimit.value = props.initialSplitLineLimit;
    } else if (localSplitLineLimit.value <= 0) {
        localSplitLineLimit.value = 500;
    }

    // subscribe to gemini api events
    EventsOn("gemini_request_start", () => {
        LogInfoRuntime("gemini request started");
        isRequestActive.value = true;
    });

    EventsOn("gemini_request_complete", () => {
        LogInfoRuntime("gemini request completed");
        isRequestActive.value = false;
        clearInterval(timerInterval.value);
    });

    EventsOn("gemini_request_canceled", () => {
        LogInfoRuntime("gemini request was canceled");
        isRequestActive.value = false;
        clearInterval(timerInterval.value);
    });
});

const shotgunGitDiffInputLines = computed(() => {
    return localShotgunGitDiffInput.value
        ? localShotgunGitDiffInput.value.split("\n").length
        : 0;
});

watch(
    () => props.initialGitDiff,
    (newVal, oldVal) => {
        if (newVal !== localShotgunGitDiffInput.value) {
            localShotgunGitDiffInput.value = newVal;
        }
    }
);

watch(
    () => props.initialSplitLineLimit,
    (newVal, oldVal) => {
        if (newVal > 0 && newVal !== localSplitLineLimit.value) {
            localSplitLineLimit.value = newVal;
        } else if (
            newVal <= 0 &&
            localSplitLineLimit.value !== 500 &&
            props.initialGitDiff === ""
        ) {
            localSplitLineLimit.value = 500;
        }
    }
);

// watch finalPrompt to keep token count updated and enforce limit
watch(
    () => props.finalPrompt,
    async (newPrompt) => {
        if (!newPrompt) {
            promptTokensCount.value = 0;
            tokenCountError.value = "";
            return;
        }
        try {
            isTokenChecking.value = true;
            const count = await CountGeminiTokens(newPrompt);
            promptTokensCount.value = count;
            tokenCountError.value = "";
            isTokenChecking.value = false;
        } catch (err) {
            // in case of error, assume over limit and keep disabled
            promptTokensCount.value = 0;
            tokenCountError.value = err?.message || "token count failed";
            isTokenChecking.value = false;
        }
    },
    { immediate: true }
);

let diffInputDebounceTimer = null;
watch(localShotgunGitDiffInput, (newVal, oldVal) => {
    clearTimeout(diffInputDebounceTimer);

    diffInputDebounceTimer = setTimeout(() => {
        if (newVal !== props.initialGitDiff) {
            emit("update:shotgunGitDiff", newVal);
        } else {
        }
        if (newVal && newVal.trim() !== "") {
            const lines = newVal.split("\n").length;
            const currentLimit = localSplitLineLimit.value;

            if (
                currentLimit === 500 ||
                (currentLimit !== lines &&
                    currentLimit ===
                        newVal
                            .substring(
                                0,
                                newVal.length -
                                    (newVal.split("\n").pop().length + 1)
                            )
                            .split("\n").length)
            ) {
                if (lines > 0 && lines !== currentLimit) {
                    localSplitLineLimit.value = lines;
                }
            } else if (lines === 0 && currentLimit !== 500) {
                localSplitLineLimit.value = 500;
            }
        } else if (
            (!newVal || newVal.trim() === "") &&
            localSplitLineLimit.value !== 500
        ) {
            localSplitLineLimit.value = 500;
        }
    }, 300);
});

let limitDebounceTimer = null;
watch(localSplitLineLimit, (newVal) => {
    clearTimeout(limitDebounceTimer);
    limitDebounceTimer = setTimeout(() => {
        if (newVal > 0 && newVal !== props.initialSplitLineLimit) {
            emit("update:splitLineLimit", newVal);
        } else if (newVal <= 0 && props.initialSplitLineLimit > 0) {
        }
    }, 300);
});

onBeforeUnmount(() => {
    // clear any pending debounced updates
    clearTimeout(diffInputDebounceTimer);
    clearTimeout(limitDebounceTimer);

    // clear timer interval if active
    if (timerInterval.value) {
        clearInterval(timerInterval.value);
    }

    // unsubscribe from events
    EventsOff("gemini_request_start");
    EventsOff("gemini_request_complete");
    EventsOff("gemini_request_canceled");

    // immediately emit the current value of localshotgungitdiffinput if it's different from the prop
    if (localShotgunGitDiffInput.value !== props.initialGitDiff) {
        emit("update:shotgunGitDiff", localShotgunGitDiffInput.value);
    } else {
    }

    // immediately emit the current value of localsplitlinelimit if it's valid and different from the prop
    if (
        localSplitLineLimit.value > 0 &&
        localSplitLineLimit.value !== props.initialSplitLineLimit
    ) {
        emit("update:splitLineLimit", localSplitLineLimit.value);
    } else {
    }
});

function handleSplitDiff() {
    if (
        !localShotgunGitDiffInput.value.trim() ||
        localSplitLineLimit.value <= 0
    ) {
        return;
    }
    emit("action", "executePromptAndSplitDiff", {
        gitDiff: localShotgunGitDiffInput.value,
        lineLimit: localSplitLineLimit.value,
    });
}

function copyDiffToClipboard() {
    if (localShotgunGitDiffInput.value) {
        navigator.clipboard
            .writeText(localShotgunGitDiffInput.value)
            .then(() => {
                copySuccess.value = true;
                // reset the success state after 2 seconds
                setTimeout(() => {
                    copySuccess.value = false;
                }, 2000);
            })
            .catch((err) => {
                LogErrorRuntime("failed to copy to clipboard: " + err);
            });
    }
}

function clearTextarea() {
    if (localShotgunGitDiffInput.value) {
        // clear the textarea
        localShotgunGitDiffInput.value = "";

        // show success message
        clearSuccess.value = true;

        // reset the success state after 2 seconds
        setTimeout(() => {
            clearSuccess.value = false;
        }, 2000);
    }
}
</script>
