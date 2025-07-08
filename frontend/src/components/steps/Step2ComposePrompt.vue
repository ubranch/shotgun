<template>
    <div class="p-4 h-full flex flex-col">
        <CustomRulesModal
            :is-visible="isPromptRulesModalVisible"
            :initial-rules="currentPromptRulesForModal"
            title="edit custom prompt rules"
            ruleType="prompt"
            @save="handleSavePromptRules"
            @cancel="handleCancelPromptRules"
        />

        <div class="flex-grow flex flex-row space-x-4 overflow-hidden">
            <div
                class="w-1/2 flex flex-col space-y-2 overflow-y-hidden px-2 py-1 border border-gray-200 dark:border-gray-700 rounded-md bg-gray-50 dark:bg-[#141414]"
            >
                <div class="flex flex-col flex-grow-[3]">
                    <label
                        for="user-task-ai"
                        class="block text-base font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >your task for ai:</label
                    >
                    <textarea
                        id="user-task-ai"
                        v-model="localUserTask"
                        class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-light-accent dark:focus:ring-dark-accent focus:border-light-accent dark:focus:border-dark-accent text-sm bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 flex-grow min-h-[100px]"
                        placeholder="describe what the ai should do..."
                    ></textarea>
                </div>

                <div class="flex flex-col flex-grow-[2]">
                    <label
                        for="rules-content"
                        class="text-base font-medium text-gray-700 dark:text-gray-300 mb-1 flex items-center"
                    >
                        custom rules:
                        <button
                            @click="openPromptRulesModal"
                            title="edit custom prompt rules"
                            class="ml-2 p-0.5 hover:bg-gray-200 dark:hover:bg-gray-700 rounded text-sm"
                        >
                            ⚙️
                        </button>
                    </label>
                    <textarea
                        id="rules-content"
                        :value="rulesContent"
                        @input="
                            (e) => emit('update:rulesContent', e.target.value)
                        "
                        class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-gray-100 dark:bg-dark-surface text-sm font-mono text-gray-900 dark:text-gray-100 flex-grow min-h-[80px]"
                        placeholder="rules for ai..."
                    ></textarea>
                </div>

                <div class="flex flex-col flex-grow-[1]">
                    <label
                        for="file-list-context"
                        class="block text-base font-medium text-gray-700 dark:text-gray-300 mb-1"
                        >files to include:</label
                    >
                    <textarea
                        id="file-list-context"
                        :value="props.fileListContext"
                        readonly
                        class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm bg-gray-100 dark:bg-dark-surface font-mono text-sm text-gray-900 dark:text-gray-100 flex-grow min-h-[50px]"
                        placeholder="file list from step 1 (prepare context) will appear here..."
                    ></textarea>
                </div>
            </div>

            <div
                class="w-1/2 flex flex-col overflow-y-auto p-2 border border-gray-200 dark:border-gray-700 rounded-md bg-white dark:bg-dark-surface"
            >
                <div class="flex justify-between items-center mb-2">
                    <div class="flex items-center space-x-2">
                        <div class="flex lg:flex-row flex-col space-x-2">
                            <button
                                v-for="(template, key) in promptTemplates"
                                :key="key"
                                @click="selectedPromptTemplateKey = key"
                                :class="[
                                    'p-2 px-3 rounded-md text-sm flex items-center',
                                    selectedPromptTemplateKey === key
                                        ? 'bg-light-accent text-white dark:bg-dark-accent'
                                        : 'bg-gray-200 text-gray-800 dark:bg-gray-700 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600'
                                ]"
                                :disabled="isLoadingFinalPrompt"
                                :title="template.name"
                            >
                                <span class="">{{ getTemplateIcon(key) }}</span>
                                <span class="font-bold">{{ getShortName(key) }}</span>
                            </button>
                        </div>
                    </div>
                    <div class="flex items-center space-x-3">
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
                            :class="['text-sm font-bold', charCountColorClass]"
                            :title="tooltipText"
                        >
                            {{ geminiTokenCount.toLocaleString() }} tokens
                        </span>
                        <button
                            @click="copyFinalPromptToClipboard"
                            :disabled="
                                !props.finalPrompt || isLoadingFinalPrompt
                            "
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
                    class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm font-mono text-sm flex-grow bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 min-h-[300px]"
                    placeholder="the final prompt will be generated here..."
                ></textarea>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from "vue";
import {
    GetCustomPromptRules,
    SetCustomPromptRules,
    CountGeminiTokens,
} from "../../../wailsjs/go/main/App";
import {
    LogInfo as LogInfoRuntime,
    LogError as LogErrorRuntime,
} from "../../../wailsjs/runtime/runtime";
import CustomRulesModal from "../CustomRulesModal.vue";

import devTemplateContentFromFile from "../../../../design/prompts/prompt_makeDiffGitFormat.md?raw";
import architectTemplateContentFromFile from "../../../../design/prompts/prompt_makePlan.md?raw";
import findBugTemplateContentFromFile from "../../../../design/prompts/prompt_analyzeBug.md?raw";
import projectManagerTemplateContentFromFile from "../../../../design/prompts/prompt_projectManager.md?raw";
import promptEnhancerTemplateContentFromFile from "../../../../design/prompts/prompt_enhancer.md?raw";

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
    architect: { name: "strategic planner and designer", content: architectTemplateContentFromFile },
    dev: { name: "builder of your plan", content: devTemplateContentFromFile },
    findBug: { name: "checker of the known & new bugs", content: findBugTemplateContentFromFile },
    projectManager: {
        name: "project scanner & analyzer of implementation",
        content: projectManagerTemplateContentFromFile,
    }
};

// helper functions for template icons and short names
function getTemplateIcon(key) {
    const icons = {
        dev: "💻",
        architect: "🏗️",
        findBug: "🐞",
        projectManager: "📋",
        promptEnhancer: "✨"
    };
    return icons[key] || "📝";
}

function getShortName(key) {
    const shortNames = {
        dev: "BUILD",
        promptEnhancer: "CREATIVE",
        architect: "PLAN",
        findBug: "Q&A",
        projectManager: "REFLECT"
    };
    return shortNames[key] || key;
}

const selectedPromptTemplateKey = ref("dev"); // default template

const isLoadingFinalPrompt = ref(false);
const copyButtonText = ref("copy");
const copySuccess = ref(false);
const geminiTokenCount = ref(0);
const isCountingTokens = ref(false);
const tokenCountError = ref("");
let tokenDebounceTimer = null;

let finalPromptDebounceTimer = null;
let userTaskInputDebounceTimer = null;

// modal state for prompt rules
const isPromptRulesModalVisible = ref(false);
const currentPromptRulesForModal = ref("");

const isFirstMount = ref(true);

const localUserTask = ref(props.userTask);

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

onMounted(async () => {
    try {
        localUserTask.value = props.userTask;
        // load rules from the backend only on the first mount
        if (isFirstMount.value) {
            const fetchedRules = await GetCustomPromptRules();
            if (!props.rulesContent) {
                emit("update:rulesContent", fetchedRules);
            }
            isFirstMount.value = false;
        }
    } catch (error) {
        console.error("failed to load custom prompt rules:", error);
        LogErrorRuntime(
            `failed to load custom prompt rules: ${error.message || error}`
        );
        if (isFirstMount.value && !props.rulesContent) {
            emit("update:rulesContent", DEFAULT_RULES);
        }
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
    }, 750);
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
    }, 500);
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

async function openPromptRulesModal() {
    try {
        currentPromptRulesForModal.value = await GetCustomPromptRules();
        isPromptRulesModalVisible.value = true;
    } catch (error) {
        console.error("error fetching prompt rules for modal:", error);
        LogErrorRuntime(
            `error fetching prompt rules for modal: ${error.message || error}`
        );
        currentPromptRulesForModal.value = props.rulesContent || DEFAULT_RULES;
        isPromptRulesModalVisible.value = true;
    }
}

async function handleSavePromptRules(newRules) {
    try {
        await SetCustomPromptRules(newRules);
        emit("update:rulesContent", newRules);
        isPromptRulesModalVisible.value = false;
        LogInfoRuntime("custom prompt rules saved successfully.");
    } catch (error) {
        console.error("error saving prompt rules:", error);
        LogErrorRuntime(`error saving prompt rules: ${error.message || error}`);
    }
}

function handleCancelPromptRules() {
    isPromptRulesModalVisible.value = false;
}

defineExpose({});
</script>
