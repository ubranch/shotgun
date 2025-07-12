<template>
    <div>
        <CustomRulesModal
            :is-visible="isCustomRulesModalVisible"
            :initial-rules="currentCustomRulesForModal"
            title="Edit Custom Ignore Rules"
            ruleType="ignore"
            @save="handleSaveCustomRules"
            @cancel="handleCancelCustomRules"
        />
        <div
            class="w-64 md:w-72 lg:w-[450px] bg-sidebar p-4 border-r border-sidebar-border flex flex-col flex-shrink-0 h-full"
        >
            <!-- project selection and file tree -->
            <div class="flex flex-col flex-grow h-full">
                <!-- select project directory button at the very top -->
                <BaseButton
                    v-if="projectRoot"
                    @click="$emit('select-directory')"
                    class="mb-4 w-full px-3 py-2 bg-sidebar-primary text-sidebar-primary-foreground text-sm font-semibold rounded-md hover:bg-sidebar-primary/90 focus:outline-none"
                >
                    open another project
                </BaseButton>

                <div class="flex justify-between items-center mb-2">
                    <h3 class="font-medium">files</h3>
                    <div class="space-x-1">
                        <BaseButton
                            @click="selectAllFiles"
                            class="text-xs px-2 py-1 bg-sidebar-primary text-sidebar-primary-foreground rounded hover:bg-sidebar-primary/90"
                        >
                            select all
                        </BaseButton>
                        <BaseButton
                            @click="deselectAllFiles"
                            class="text-xs px-2 py-1"
                        >
                            deselect all
                        </BaseButton>
                        <BaseButton
                            @click="handleEditCustomRules"
                            class="text-xs px-2 py-1"
                        >
                            edit rules
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
                    <div v-else-if="loadingError" class="p-3 text-destructive">
                        error loading files: {{ loadingError }}
                    </div>
                    <div v-else class="p-3">Loading files...</div>
                </div>

                <div class="mt-2 flex justify-around gap-4">
                    <label class="flex items-center text-sm">
                        <input
                            type="checkbox"
                            :checked="useGitignore"
                            @change="
                                $emit('toggle-gitignore', $event.target.checked)
                            "
                            class="form-checkbox h-4 w-4 text-sidebar-primary rounded border-border focus:ring-sidebar-primary mr-2"
                        />
                        use .gitignore rules
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
                        use custom rules
                    </label>
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
</script>

<style scoped>
/* add your styles here */
</style>
