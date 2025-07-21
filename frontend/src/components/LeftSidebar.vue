<template>
    <div>
        <CustomRulesModal
            :is-visible="isCustomRulesModalVisible"
            :initial-rules="currentCustomRulesForModal"
            title="edit custom ignore rules"
            ruleType="ignore"
            @save="handleSaveCustomRules"
            @cancel="handleCancelCustomRules"
        />
        <CustomRulesModal
            :is-visible="isPromptRulesModalVisible"
            :initial-rules="currentPromptRulesForModal_prompt"
            title="edit custom prompt rules"
            ruleType="prompt"
            @save="handleSavePromptRules_prompt"
            @cancel="handleCancelPromptRules_prompt"
        />
        <div
            class="sidebar-container flex item-top h-full"
            :class="{ collapsed: isSidebarCollapsed }"
        >
            <div
                class="sidebar-content w-64 lg:w-[450px] bg-sidebar p-4 border-r border-sidebar-border flex flex-col flex-shrink-0 h-full max-[900px]:w-[415px]"
            >
                <!-- project selection and file tree -->
                <div class="flex flex-col flex-grow h-full">
                    <!-- project actions: open project & reset -->
                    <div
                        v-if="projectRoot"
                        class="mb-4 flex items-center gap-2"
                    >
                        <BaseButton
                            @click="handleReset"
                            :title="'reset application'"
                            variant="danger"
                            class="aspect-square text-base"
                        >
                            <!-- simple refresh icon -->
                            <span class="text-base"> reset </span>
                        </BaseButton>
                        <BaseButton
                            @click="$emit('select-directory')"
                            class="flex-1 px-3 py-2 bg-sidebar-primary text-sidebar-primary-foreground text-base font-semibold rounded-md hover:bg-sidebar-primary/90 focus:outline-none"
                        >
                            <span class="text-base"> open another project </span>
                        </BaseButton>
                        <BaseButton
                            @click="openPromptRulesModal_prompt"
                            title="edit custom prompt rules"
                            class="px-2 py-1"
                        >
                            <span class="text-base"> rules </span>
                        </BaseButton>
                    </div>

                    <div
                        class="flex flex-row justify-between items-center mb-2"
                    >
                        <div class="space-x-1">
                            <BaseButton
                                @click="selectAllFiles"
                                class="text-xs px-2 py-1 bg-sidebar-primary text-sidebar-primary-foreground rounded hover:bg-sidebar-primary/90"
                            >
                                <span class="text-base"> select all </span>
                            </BaseButton>
                            <BaseButton
                                @click="deselectAllFiles"
                                class="text-xs px-2 py-1"
                            >
                                <span class="text-base"> deselect all </span>
                            </BaseButton>
                            <BaseButton
                                @click="resetFileSelections"
                                variant="warning"
                                class="text-xs px-2 py-1"
                            >
                                <span class="text-base"> default </span>
                            </BaseButton>
                            <BaseButton
                                @click="handleRefreshProject"
                                :disabled="isRefreshing"
                                class="text-xs px-2 py-1 bg-sidebar-primary text-sidebar-primary-foreground rounded hover:bg-sidebar-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                                title="refresh project files"
                            >
                                <span class="text-base"> refresh </span>
                            </BaseButton>
                        </div>
                    </div>

                    <!-- file tree -->
                    <div
                        class="border border-border rounded min-h-0 bg-card text-sm overflow-auto flex-grow h-0"
                    >
                        <FileTree
                            v-if="fileTreeNodes && fileTreeNodes.length > 0"
                            :nodes="fileTreeNodes"
                            :loading-error="loadingError"
                            @toggle-exclude="
                                (path) => $emit('toggle-exclude', path)
                            "
                            @add-log="(log) => $emit('add-log', log)"
                        />
                        <div v-else-if="!projectRoot" class="p-3">
                            select a project folder to view files.
                        </div>
                        <div
                            v-else-if="loadingError"
                            class="p-3 text-destructive"
                        >
                            error loading files: {{ loadingError }}
                        </div>
                        <div v-else class="p-3">loading files...</div>
                    </div>

                    <div class="mt-2 flex justify-around gap-4">
                        <label class="flex items-center text-sm">
                            <input
                                type="checkbox"
                                :checked="useGitignore"
                                @change="
                                    $emit(
                                        'toggle-gitignore',
                                        $event.target.checked
                                    )
                                "
                                class="form-checkbox h-4 w-4 text-sidebar-primary rounded border-border focus:ring-sidebar-primary mr-2"
                            />
                            <span class="text-base"> use .gitignore rules </span>
                        </label>
                        <label class="flex items-center text-sm">
                            <input
                                type="checkbox"
                                :checked="useCustomIgnore"
                                @change="
                                    $emit(
                                        'toggle-custom-ignore',
                                        $event.target.checked
                                    )
                                "
                                class="form-checkbox h-4 w-4 text-sidebar-primary rounded border-border focus:ring-sidebar-primary mr-2"
                            />
                            <span class="text-base"> use custom rules </span>
                        </label>
                    </div>
                </div>
            </div>

            <!-- collapse toggle button - moved to the right edge -->
            <div
                class="absolute bottom-[422px] -right-4 z-10 collapse-toggle flex items-center justify-center cursor-pointer bg-sidebar-primary hover:bg-sidebar-primary/90"
                @click="toggleSidebar"
            >
                <div class="flex items-center justify-center h-8 w-8">
                    <i
                        class="arrow-icon"
                        :class="{
                            'arrow-left': !isSidebarCollapsed,
                            'arrow-right': isSidebarCollapsed,
                        }"
                    ></i>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { defineProps, defineEmits, ref } from "vue";
import FileTree from "./FileTree.vue"; // import the existing filetree
import CustomRulesModal from "./CustomRulesModal.vue";
import BaseButton from "./BaseButton.vue";
import {
    GetCustomIgnoreRules,
    SetCustomIgnoreRules,
    GetCustomPromptRules,
    SetCustomPromptRules,
} from "../../wailsjs/go/main/App";
import {
    LogError as LogErrorRuntime,
    LogInfo as LogInfoRuntime,
} from "../../wailsjs/runtime/runtime";

/**
 * props for leftsidebar:
 * - usegitignore: enables .gitignore rules for file parsing
 * - usecustomignore: enables custom ignore.glob rules for file parsing
 */
const props = defineProps({
    currentStep: { type: Number, required: true },
    steps: { type: Array, required: true }, // array of { id: number, title: string, completed: boolean }
    projectRoot: { type: String, default: "" },
    fileTreeNodes: { type: Array, default: () => [] },
    useGitignore: { type: Boolean, default: true },
    useCustomIgnore: { type: Boolean, default: false },
    loadingError: { type: String, default: "" },
    isRefreshing: { type: Boolean, default: false },
});

const emit = defineEmits([
    "navigate",
    "toggle-gitignore",
    "toggle-custom-ignore",
    "toggle-exclude",
    "custom-rules-updated",
    "add-log",
    "select-all-files",
    "deselect-all-files",
    "reset-file-selections",
    "select-directory",
    "sidebar-toggle",
    "reset",
    "update:rulesContent",
    "refresh-project",
]);

const isCustomRulesModalVisible = ref(false);
const currentCustomRulesForModal = ref("");
const isSidebarCollapsed = ref(false);

// state for prompt rules modal
const isPromptRulesModalVisible = ref(false);
const currentPromptRulesForModal_prompt = ref("");

function toggleSidebar() {
    isSidebarCollapsed.value = !isSidebarCollapsed.value;
    emit("sidebar-toggle", isSidebarCollapsed.value);
}

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
        // show a placeholder or error message in the textarea if loading fails
        currentCustomRulesForModal.value =
            "# error loading rules. please check application logs.\n# you can still edit and save.";
        isCustomRulesModalVisible.value = true; // still open modal
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
        emit("custom-rules-updated"); // notify mainlayout to refresh
    } catch (error) {
        console.error("error saving custom ignore rules:", error);
        LogErrorRuntime(`error saving custom rules: ${error.message || error}`);
        emit("add-log", {
            message: `failed to save custom rules: ${error.message || error}`,
            type: "error",
        });
        // keep modal open for user to retry or copy content, or show an error in the modal itself.
    }
}

function handleCancelCustomRules() {
    isCustomRulesModalVisible.value = false;
}

async function openPromptRulesModal_prompt() {
    try {
        currentPromptRulesForModal_prompt.value = await GetCustomPromptRules();
        isPromptRulesModalVisible.value = true;
    } catch (error) {
        console.error("error fetching prompt rules for modal:", error);
        LogErrorRuntime(
            `error fetching prompt rules for modal: ${error.message || error}`
        );
        emit("add-log", {
            message: `failed to load prompt rules: ${error.message || error}`,
            type: "error",
        });
        currentPromptRulesForModal_prompt.value = "# error loading rules.";
        isPromptRulesModalVisible.value = true;
    }
}

async function handleSavePromptRules_prompt(newRules) {
    try {
        await SetCustomPromptRules(newRules);
        isPromptRulesModalVisible.value = false;
        LogInfoRuntime("custom prompt rules saved successfully.");
        emit("update:rulesContent", newRules);
    } catch (error) {
        console.error("error saving prompt rules:", error);
        LogErrorRuntime(`error saving prompt rules: ${error.message || error}`);
        // optionally, emit a log to the user
    }
}

function handleCancelPromptRules_prompt() {
    isPromptRulesModalVisible.value = false;
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
    ); // allow any if all completed
}

function selectAllFiles() {
    emit("select-all-files");
    emit("add-log", {
        message: "selecting all files",
        type: "info",
    });
}

function deselectAllFiles() {
    emit("deselect-all-files");
    emit("add-log", {
        message: "deselecting all files",
        type: "info",
    });
}

function resetFileSelections() {
    emit("reset-file-selections");
    emit("add-log", {
        message: "resetting file selections to default",
        type: "info",
    });
}

function handleToggleExclude(node) {
    console.log(
        `DEBUG: LeftSidebar received toggle-exclude for node: ${node.name}, path: ${node.relPath}`
    );
    emit("toggle-exclude", node);
}

function handleEditCustomRules() {
    openCustomRulesModal();
}

function handleReset() {
    emit("reset");
    emit("add-log", {
        message: "resetting application state",
        type: "info",
    });
}

function handleRefreshProject() {
    emit("refresh-project");
    emit("add-log", {
        message: "refreshing project files",
        type: "info",
    });
}
</script>

<style scoped>
.sidebar-container {
    position: relative;
}

.sidebar-container.collapsed .sidebar-content {
    width: 0 !important;
    padding: 0;
    overflow: hidden;
}

.collapse-toggle {
    width: 16px;
    border-top-right-radius: 4px;
    border-bottom-right-radius: 4px;
    border-right: 1px solid var(--border);
    border-top: 1px solid var(--border);
    border-bottom: 1px solid var(--border);
    color: var(--sidebar-primary-foreground);
    box-shadow: 2px 0 5px rgba(0, 0, 0, 0.1);
    margin-top: 16px;
    height: 40px;
}

.arrow-icon {
    width: 0;
    height: 0;
    border-top: 6px solid transparent;
    border-bottom: 6px solid transparent;
}

.arrow-left {
    border-right: 6px solid currentColor;
}

.arrow-right {
    border-left: 6px solid currentColor;
}
</style>
