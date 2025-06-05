<template>
    <CustomRulesModal
        :is-visible="isCustomRulesModalVisible"
        :initial-rules="currentCustomRulesForModal"
        title="edit custom ignore rules"
        ruleType="ignore"
        @save="handleSaveCustomRules"
        @cancel="handleCancelCustomRules"
    />
    <aside
        class="w-64 md:w-72 lg:w-80 bg-light-bg dark:bg-dark-surface p-4 border-r border-light-border dark:border-dark-border flex flex-col flex-shrink-0 h-full"
    >
        <!-- Project Selection and File Tree -->
        <div class="flex flex-col flex-grow h-full">
            <button
                @click="$emit('select-folder')"
                class="w-full px-4 py-2 mb-2 bg-light-accent dark:bg-dark-accent text-white font-semibold rounded-md hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50"
            >
                select project folder
            </button>
            <div
                v-if="projectRoot"
                class="text-xs text-gray-600 dark:text-gray-400 mb-2 break-all"
            >
                selected: {{ projectRoot }}
            </div>

            <div v-if="projectRoot" class="mb-2">
                <label
                    class="flex items-center text-sm text-gray-700 dark:text-gray-300"
                    title="uses .gitignore file if present in the project folder"
                >
                    <input
                        type="checkbox"
                        :checked="useGitignore"
                        @change="
                            $emit('toggle-gitignore', $event.target.checked)
                        "
                        class="form-checkbox h-4 w-4 text-blue-600 rounded border-gray-300 dark:border-gray-600 focus:ring-blue-500 mr-2"
                    />
                    use .gitignore rules
                </label>
                <label
                    class="flex items-center text-sm text-gray-700 dark:text-gray-300 mt-1"
                    title="uses ignore.glob file if present in the project folder"
                >
                    <input
                        type="checkbox"
                        :checked="useCustomIgnore"
                        @change="
                            $emit('toggle-custom-ignore', $event.target.checked)
                        "
                        class="form-checkbox h-4 w-4 text-indigo-600 rounded border-gray-300 dark:border-gray-600 focus:ring-indigo-500 mr-2"
                    />
                    use custom rules
                    <button
                        @click="openCustomRulesModal"
                        title="edit custom ignore rules"
                        class="ml-2 p-0.5 hover:bg-gray-200 dark:hover:bg-gray-700 rounded text-xs"
                    >
                        ⚙️
                    </button>
                </label>
            </div>

            <h2
                class="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-2"
            >
                project files
            </h2>
            <div
                class="border border-gray-300 dark:border-gray-600 rounded min-h-0 bg-white dark:bg-dark-surface text-sm overflow-auto flex-grow h-0"
            >
                <FileTree
                    v-if="fileTreeNodes && fileTreeNodes.length"
                    :nodes="fileTreeNodes"
                    :project-root="projectRoot"
                    @toggle-exclude="(node) => $emit('toggle-exclude', node)"
                />
                <p
                    v-else-if="projectRoot && !loadingError"
                    class="p-2 text-xs text-gray-500 dark:text-gray-400"
                >
                    loading tree...
                </p>
                <p
                    v-else-if="!projectRoot"
                    class="p-2 text-xs text-gray-500 dark:text-gray-400"
                >
                    select a project folder to see files.
                </p>
                <p v-if="loadingError" class="p-2 text-xs text-red-500">
                    {{ loadingError }}
                </p>
            </div>
        </div>
    </aside>
</template>

<script setup>
import { defineProps, defineEmits, ref } from "vue";
import FileTree from "./FileTree.vue"; // Import the existing FileTree
import CustomRulesModal from "./CustomRulesModal.vue";
import {
    GetCustomIgnoreRules,
    SetCustomIgnoreRules,
} from "../../wailsjs/go/main/App";
import {
    LogError as LogErrorRuntime,
    LogInfo as LogInfoRuntime,
} from "../../wailsjs/runtime/runtime";

/**
 * Props for LeftSidebar:
 * - useGitignore: enables .gitignore rules for file parsing
 * - useCustomIgnore: enables custom ignore.glob rules for file parsing
 */
const props = defineProps({
    currentStep: { type: Number, required: true },
    steps: { type: Array, required: true }, // Array of { id: Number, title: String, completed: Boolean }
    projectRoot: { type: String, default: "" },
    fileTreeNodes: { type: Array, default: () => [] },
    useGitignore: { type: Boolean, default: true },
    useCustomIgnore: { type: Boolean, default: false },
    loadingError: { type: String, default: "" },
});

const emit = defineEmits([
    "navigate",
    "select-folder",
    "toggle-gitignore",
    "toggle-custom-ignore",
    "toggle-exclude",
    "custom-rules-updated",
    "add-log",
]);

const isCustomRulesModalVisible = ref(false);
const currentCustomRulesForModal = ref("");

async function openCustomRulesModal() {
    try {
        currentCustomRulesForModal.value = await GetCustomIgnoreRules();
        isCustomRulesModalVisible.value = true;
    } catch (error) {
        console.error("error fetching custom ignore rules:", error);
        LogErrorRuntime(
            `error fetching custom rules: ${error.message || error}`
        );
        emit("add-log", {
            message: `failed to load custom rules: ${error.message || error}`,
            type: "error",
        });
        // Show a placeholder or error message in the textarea if loading fails
        currentCustomRulesForModal.value =
            "# error loading rules. please check application logs.\n# you can still edit and save.";
        isCustomRulesModalVisible.value = true; // Still open modal
    }
}

async function handleSaveCustomRules(newRules) {
    try {
        await SetCustomIgnoreRules(newRules);
        isCustomRulesModalVisible.value = false;
        LogInfoRuntime(
            "custom ignore rules saved successfully via leftsidebar."
        );
        emit("add-log", {
            message: "custom ignore rules saved.",
            type: "success",
        });
        emit("custom-rules-updated"); // Notify MainLayout to refresh
    } catch (error) {
        console.error("error saving custom ignore rules:", error);
        LogErrorRuntime(`error saving custom rules: ${error.message || error}`);
        emit("add-log", {
            message: `failed to save custom rules: ${error.message || error}`,
            type: "error",
        });
        // Keep modal open for user to retry or copy content, or show an error in the modal itself.
    }
}

function handleCancelCustomRules() {
    isCustomRulesModalVisible.value = false;
}

function canNavigateToStep(stepId) {
    if (stepId === props.currentStep) return true;
    const targetStep = props.steps.find((s) => s.id === stepId);
    if (targetStep && targetStep.completed) return true;
    const firstUncompletedStep = props.steps.find((s) => !s.completed);
    const firstUncompletedStepId = firstUncompletedStep
        ? firstUncompletedStep.id
        : undefined;
    return (
        stepId === firstUncompletedStepId ||
        (firstUncompletedStepId === undefined && targetStep)
    ); // Allow any if all completed
}
</script>

<style scoped>
/* Add your styles here */
</style>
