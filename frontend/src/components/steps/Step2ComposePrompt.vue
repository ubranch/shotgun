<template>
    <div class="p-4 h-full flex flex-col">
        <!-- error display for context generation failures -->
        <div
            v-if="isErrorContext"
            class="mb-4 p-4 border border-red-300 dark:border-red-700 rounded-lg bg-red-50 dark:bg-red-900 dark:bg-opacity-20 shadow-sm"
        >
            <h4
                class="text-lg font-semibold mb-2 text-red-600 dark:text-red-400"
            >
                context generation error
            </h4>
            <pre
                class="text-sm whitespace-pre-wrap bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 p-3 border border-red-200 dark:border-red-700 rounded-md overflow-auto max-h-[150px]"
                >{{ errorMessage }}</pre
            >
            <p class="mt-3 text-sm text-red-600 dark:text-red-400">
                go back to step 1 to reduce the project scope by excluding more
                files or using a smaller project
            </p>
        </div>

        <!-- custom rules modal removed per user request -->
        <div class="flex-grow flex flex-row space-x-0 overflow-hidden">
            <!--
                group-[.sidebar-open]/layout:max-[900px]:hidden:
                hides the user query panel when the sidebar is open and screen width is ≤ 900px
            -->
            <div
                class="w-3/5 group-[.sidebar-open]/layout:w-2/5 flex flex-col space-y-2 overflow-y-hidden px-2 py-2 border border-accent rounded-md bg-white dark:bg-dark-surface mr-2 group-[.sidebar-open]/layout:max-[900px]:hidden"
            >
                <div class="flex flex-col flex-grow-[3]">
                    <!-- <label
                        for="user-task-ai"
                        class="block text-base font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >your query for ai:</label
                    > -->
                    <textarea
                        id="user-task-ai"
                        v-model="localUserTask"
                        spellcheck="false"
                        class="w-full p-2 border border-accent rounded-md shadow-sm focus:ring-light-accent dark:focus:ring-dark-accent focus:border-light-accent dark:focus:border-dark-accent text-sm bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 flex-grow min-h-[100px]"
                        placeholder="describe what the ai should do..."
                    ></textarea>
                </div>

                <!-- custom rules textarea commented out per user request -->
                <!-- <div class="flex flex-col flex-grow-[1]">
                    <label
                        for="file-list-context"
                        class="text-base font-medium text-gray-700 dark:text-gray-300 mb-1 flex items-center"
                    >
                        file list context:
                    </label>
                    <textarea
                        id="file-list-context"
                        :value="props.fileListContext"
                        readonly
                        spellcheck="false"
                        class="w-full p-2 border border-accent rounded-md shadow-sm bg-gray-100 dark:bg-dark-surface font-mono text-sm text-gray-900 dark:text-gray-100 flex-grow min-h-[50px]"
                        placeholder="file list from step 1 (prepare context) will appear here..."
                    ></textarea>
                </div> -->
            </div>

            <!--
                group-[.sidebar-open]/layout:max-[900px]:w-full:
                expands the final prompt panel to full width when the sidebar is open and screen width is ≤ 900px
            -->
            <div
                class="w-3/5 flex flex-col overflow-y-auto p-2 border border-accent rounded-md bg-white dark:bg-dark-surface group-[.sidebar-open]/layout:max-[900px]:w-full"
            >
                <div class="flex justify-between items-center mb-2">
                    <div class="flex items-center space-x-2">
                        <div class="flex flex-row space-x-2">
                            <!-- mode selector: buttons for large screens, dropdown for small screens -->
                            <template v-if="!isSmallScreen">
                                <BaseButton
                                    v-for="(template, key) in promptTemplates"
                                    :key="key"
                                    @click="selectedPromptTemplateKey = key"
                                    :class="[
                                        'p-2 px-3 rounded-md text-sm flex items-center font-semibold hover:bg-sidebar-primary/90 focus:outline-none',
                                        selectedPromptTemplateKey === key
                                            ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                                            : 'text-gray-200 hover:text-white',
                                    ]"
                                    :disabled="isLoadingFinalPrompt"
                                    :title="template.name"
                                >
                                    <span class="font-bold">{{
                                        getShortName(key)
                                    }}</span>
                                </BaseButton>
                            </template>
                            <template v-else>
                                <div class="relative inline-block w-full">
                                    <select
                                        v-model="selectedPromptTemplateKey"
                                        :disabled="isLoadingFinalPrompt"
                                        class="appearance-none pr-8 p-2 rounded-md text-sm flex items-center border-2 font-semibold focus:outline-none border-border bg-background focus-visible:ring-primary text-gray-200 w-full"
                                    >
                                        <option
                                            v-for="(
                                                template, key
                                            ) in promptTemplates"
                                            :key="key"
                                            :value="key"
                                        >
                                            {{ getShortName(key) }}
                                        </option>
                                    </select>
                                    <!-- chevron arrow icon -->
                                    <svg
                                        class="absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none h-4 w-4 text-gray-500 dark:text-gray-400"
                                        xmlns="http://www.w3.org/2000/svg"
                                        viewBox="0 0 20 20"
                                        fill="currentColor"
                                    >
                                        <path
                                            fill-rule="evenodd"
                                            d="M5.23 7.21a.75.75 0 011.06.02L10 10.92l3.71-3.69a.75.75 0 111.06 1.06l-4.24 4.25a.75.75 0 01-1.06 0L5.23 8.29a.75.75 0 01.02-1.08z"
                                            clip-rule="evenodd"
                                        />
                                    </svg>
                                </div>
                            </template>
                        </div>
                    </div>
                    <!-- <div class="flex flex-row space-x-2">
                        <span
                            v-if="isCountingTokens"
                            class="text-sm text-gray-500"
                        >
                            counting...
                        </span>
                        <span
                            v-else-if="tokenCountError"
                            class="text-sm text-red-500"
                            :title="tokenCountError"
                        >
                            error
                        </span>
                        <span
                            v-else
                            :class="[
                                'text-sm font-bold px-2 py-1 ml-2 rounded-xl',
                                charCountColorClass === 'text-green-600'
                                    ? 'bg-green-100 dark:bg-green-900/30'
                                    : charCountColorClass === 'text-yellow-500'
                                      ? 'bg-yellow-100 dark:bg-yellow-900/30'
                                      : 'bg-red-100 dark:bg-red-900/30',
                            ]"
                            :title="tooltipText"
                        >
                            {{ geminiTokenCount.toLocaleString() }}
                        </span>
                    </div> -->
                    <div class="flex items-center space-x-3">
                        <!-- refresh button -->
                        <BaseButton
                            @click="refreshPrompt"
                            :disabled="isLoadingFinalPrompt"
                            :class="[
                                'p-2 px-3 rounded-md text-sm flex items-center font-semibold hover:bg-sidebar-primary/90 focus:outline-none',
                                refreshing
                                    ? 'bg-sidebar-primary text-white'
                                    : 'text-gray-200 hover:text-white',
                            ]"
                            title="regenerate prompt"
                        >
                            <span class="text-base">update</span>
                        </BaseButton>
                        <BaseButton
                            @click="copyFinalPromptToClipboard"
                            :disabled="
                                !props.finalPrompt || isLoadingFinalPrompt
                            "
                            class="px-3 py-2 bg-sidebar-primary text-sidebar-primary-foreground text-base font-semibold rounded-md hover:bg-sidebar-primary/90 focus:outline-none disabled:bg-gray-300 dark:disabled:bg-gray-700 flex items-center gap-1"
                            :class="{
                                'bg-green-600 dark:bg-green-700': copySuccess,
                            }"
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
                            <span class="text-base">{{ copyButtonText }}</span>
                        </BaseButton>
                    </div>
                </div>

                <div
                    v-if="isLoadingFinalPrompt"
                    class="flex-grow flex justify-center items-center"
                >
                    <div
                        class="animate-spin rounded-full h-8 w-8 border-b-2 border-light-accent dark:border-dark-accent"
                    ></div>
                    <p class="text-gray-500 dark:text-gray-400 ml-2">
                        updating prompt...
                    </p>
                </div>

                <textarea
                    v-else
                    :value="props.finalPrompt"
                    @input="(e) => emit('update:finalPrompt', e.target.value)"
                    spellcheck="false"
                    class="w-full p-2 border border-accent rounded-md shadow-sm font-mono text-sm flex-grow bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 min-h-[300px]"
                    placeholder="the final prompt will be generated here..."
                ></textarea>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted, computed, onUnmounted } from "vue";
import { CountGeminiTokens } from "../../../wailsjs/go/main/App";
import {
    LogInfo as LogInfoRuntime,
    LogError as LogErrorRuntime,
} from "../../../wailsjs/runtime/runtime";
import BaseButton from "../BaseButton.vue";

import devTemplateContentFromFile from "../../../../design/prompts/prompt_makeDiffGitFormat_v2.md?raw";
import architectTemplateContentFromFile from "../../../../design/prompts/prompt_makePlan_v2.md?raw";
import findBugTemplateContentFromFile from "../../../../design/prompts/prompt_analyzeBug_v2.md?raw";
import projectManagerTemplateContentFromFile from "../../../../design/prompts/prompt_projectManager_v2.md?raw";
import promptEnhancerTemplateContentFromFile from "../../../../design/prompts/prompt_enhancer_v2.md?raw";

const props = defineProps({
    fileListContext: {
        type: String,
        default: "",
    },
    platform: {
        // to know if we are on macos
        type: String,
        default: "unknown",
    },
    userTask: {
        type: String,
        default: "",
    },
    rulesContent: {
        type: String,
        default: "",
    },
    finalPrompt: {
        type: String,
        default: "",
    },
});

const emit = defineEmits([
    "update:finalPrompt",
    "update:userTask",
    "update:rulesContent",
]);

const promptTemplates = {
    promptEnhancer: {
        name: "prompt engineer of your task",
        content: promptEnhancerTemplateContentFromFile,
    },
    architect: {
        name: "strategic planner and designer",
        content: architectTemplateContentFromFile,
    },
    dev: { name: "builder of your plan", content: devTemplateContentFromFile },
    findBug: {
        name: "checker of the known & new bugs",
        content: findBugTemplateContentFromFile,
    },
    projectManager: {
        name: "project scanner & analyzer of implementation",
        content: projectManagerTemplateContentFromFile,
    },
};

// helper functions for template icons and short names
function getTemplateIcon(key) {
    const icons = {
        dev: "",
        architect: "",
        findBug: "",
        projectManager: "",
        promptEnhancer: "",
    };
    return icons[key] || "";
}

function getShortName(key) {
    const shortNames = {
        dev: "BUILD",
        promptEnhancer: "PROMPT",
        architect: "PLAN",
        findBug: "BUG",
        projectManager: "REFLECT",
    };
    return shortNames[key] || key;
}

const selectedPromptTemplateKey = ref(Object.keys(promptTemplates)[0]); // default to first template

const isLoadingFinalPrompt = ref(false);
const copyButtonText = ref("copy");
const copySuccess = ref(false);
const geminiTokenCount = ref(0);
const isCountingTokens = ref(false);
const tokenCountError = ref("");
let tokenDebounceTimer = null;

// refresh button state
const refreshing = ref(false);

let finalPromptDebounceTimer = null;
let userTaskInputDebounceTimer = null;

// modal state for prompt rules removed
// const isPromptRulesModalVisible = ref(false);
// const currentPromptRulesForModal = ref("");

const isFirstMount = ref(true);

const localUserTask = ref(props.userTask);

// Error detection (same logic as Step 1)
const isErrorContext = computed(() => {
    if (!props.fileListContext) return false;
    // consider only the first non-blank line to decide if the backend sent an error
    const firstLine = props.fileListContext
        .trimStart()
        .split("\n", 1)[0]
        .toLowerCase();
    return firstLine.startsWith("error:");
});

const errorMessage = computed(() => {
    if (!isErrorContext.value || !props.fileListContext) return "";

    // check if starts with "Error:" (case insensitive)
    const lowerCaseContext = props.fileListContext.toLowerCase();
    if (lowerCaseContext.startsWith("error:")) {
        return props.fileListContext
            .substring(props.fileListContext.indexOf(":") + 1)
            .trim();
    }

    // if it contains "error:" elsewhere, try to extract the message
    if (lowerCaseContext.includes("error:")) {
        const errorIndex = lowerCaseContext.indexOf("error:");
        return props.fileListContext.substring(errorIndex).trim();
    }

    return props.fileListContext.trim();
});

// character count and related computed properties
const charCount = computed(() => {
    return (props.finalPrompt || "").length;
});

const charCountColorClass = computed(() => {
    const count = geminiTokenCount.value;
    if (count < 1000000) {
        return "text-green-600";
    } else if (count <= 4000000) {
        return "text-yellow-500"; // using 500 for better visibility on white bg
    } else {
        return "text-red-600";
    }
});

const tooltipText = computed(() => {
    if (isCountingTokens.value) return "calculating tokens...";
    if (tokenCountError.value) return `error: ${tokenCountError.value}`;

    return `prompt contains ${geminiTokenCount.value.toLocaleString()} gemini tokens`;
});

const DEFAULT_RULES = `no additional rules`;

// responsive: detect narrow screens (< 900px)
const isSmallScreen = ref(window.innerWidth < 1200);

function updateScreenSize() {
    isSmallScreen.value = window.innerWidth < 1200;
}

onMounted(() => {
    window.addEventListener("resize", updateScreenSize);
});

onUnmounted(() => {
    window.removeEventListener("resize", updateScreenSize);
});

onMounted(async () => {
    try {
        localUserTask.value = props.userTask;
        // removed: load rules from the backend only on the first mount
        // if (isFirstMount.value) {
        //     const fetchedRules = await GetCustomPromptRules();
        //     if (!props.rulesContent) {
        //         emit("update:rulesContent", fetchedRules);
        //     }
        //     isFirstMount.value = false;
        // }
    } catch (error) {
        console.error("failed to load custom prompt rules:", error);
        LogErrorRuntime(
            `failed to load custom prompt rules: ${error.message || error}`
        );
        // if (isFirstMount.value && !props.rulesContent) {
        //     emit("update:rulesContent", DEFAULT_RULES);
        // }
        isFirstMount.value = false;
    }

    // always generate initial prompt if not already available
    // this ensures token calculation triggers even when file list context or user task are initially empty
    if (!props.finalPrompt) {
        debouncedUpdateFinalPrompt();
    }
});

async function updateFinalPrompt() {
    isLoadingFinalPrompt.value = true;
    await new Promise((resolve) => setTimeout(resolve, 100));

    const currentTemplateContent =
        promptTemplates[selectedPromptTemplateKey.value].content;
    let populatedPrompt = currentTemplateContent;
    populatedPrompt = populatedPrompt.replace(
        "{TASK}",
        props.userTask || "no task provided by the user."
    );
    populatedPrompt = populatedPrompt.replace("{RULES}", props.rulesContent);
    populatedPrompt = populatedPrompt.replace(
        "{FILE_STRUCTURE}",
        props.fileListContext || "no file structure context provided."
    );

    // insert current date in yyyy-mm-dd format
    const now = new Date();
    const yyyy = now.getFullYear();
    const mm = String(now.getMonth() + 1).padStart(2, "0");
    const dd = String(now.getDate()).padStart(2, "0");
    const currentDate = `${yyyy}-${mm}-${dd}`;
    populatedPrompt = populatedPrompt.replaceAll("{CURRENT_DATE}", currentDate);

    emit("update:finalPrompt", populatedPrompt);

    // trigger token counting immediately for the freshly generated prompt
    countTokensForPrompt(populatedPrompt);

    isLoadingFinalPrompt.value = false;
}

function debouncedUpdateFinalPrompt() {
    clearTimeout(finalPromptDebounceTimer);
    finalPromptDebounceTimer = setTimeout(() => {
        updateFinalPrompt();
    }, 300);
}

// refresh prompt functionality
function refreshPrompt() {
    if (isLoadingFinalPrompt.value) return;

    // visual feedback
    refreshing.value = true;

    // force prompt regeneration
    updateFinalPrompt();

    // reset refreshing state after a short delay
    setTimeout(() => {
        refreshing.value = false;
    }, 500);
}

watch(
    () => props.userTask,
    (newValue) => {
        if (newValue !== localUserTask.value) {
            localUserTask.value = newValue;
        }
    }
);

watch(localUserTask, (currentValue) => {
    clearTimeout(userTaskInputDebounceTimer);
    userTaskInputDebounceTimer = setTimeout(() => {
        if (currentValue !== props.userTask) {
            emit("update:userTask", currentValue);
        }
    }, 300);
});

watch(
    [
        () => props.userTask,
        () => props.rulesContent,
        () => props.fileListContext,
        selectedPromptTemplateKey,
    ],
    () => {
        debouncedUpdateFinalPrompt();
    },
    { deep: true }
);

watch(selectedPromptTemplateKey, () => {
    LogInfoRuntime(
        `prompt template changed to: ${
            promptTemplates[selectedPromptTemplateKey.value].name
        }. updating final prompt.`
    );
    debouncedUpdateFinalPrompt();
});

const countTokensForPrompt = (prompt) => {
    clearTimeout(tokenDebounceTimer);
    if (!prompt) {
        geminiTokenCount.value = 0;
        tokenCountError.value = "";
        return;
    }
    isCountingTokens.value = true;
    tokenCountError.value = "";
    tokenDebounceTimer = setTimeout(async () => {
        try {
            const count = await CountGeminiTokens(prompt);
            geminiTokenCount.value = count;
        } catch (err) {
            console.error("token counting error:", err);
            tokenCountError.value = err.message || "token count failed";
            geminiTokenCount.value = 0;
        } finally {
            isCountingTokens.value = false;
        }
    }, 200);
};

watch(
    () => props.finalPrompt,
    (newPrompt) => {
        countTokensForPrompt(newPrompt);
    },
    { immediate: true }
);

async function copyFinalPromptToClipboard() {
    if (!props.finalPrompt) return;
    try {
        await navigator.clipboard.writeText(props.finalPrompt);
        copyButtonText.value = "copied!";
        copySuccess.value = true;
        setTimeout(() => {
            copyButtonText.value = "copy";
            copySuccess.value = false;
        }, 2000);
    } catch (err) {
        console.error("failed to copy final prompt: ", err);
        if (props.platform === "darwin" && err) {
            console.error(
                "darvin clipboardsettext failed for final prompt:",
                err
            );
        }
        copyButtonText.value = "failed!";
        copySuccess.value = false;
        setTimeout(() => {
            copyButtonText.value = "copy";
        }, 2000);
    }
}

// removed: openPromptRulesModal, handleSavePromptRules, handleCancelPromptRules

defineExpose({});
</script>
