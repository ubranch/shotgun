<template>
    <div
        class="relative p-4 h-full flex flex-col"
        style="--wails-drop-target: drop;"
        @dragenter.prevent="onDragEnter"
        @dragover.prevent="onDragOver"
        @dragleave.prevent="onDragLeave"
        @drop.prevent="onDrop"
    >
        <!-- drag-drop overlay disabled (no drag-n-drop after a project is open) -->
        <div
            v-if="false"
            class="absolute inset-4 z-10 flex flex-col justify-center items-center border-2 border-dashed rounded-lg bg-light-accent/10 dark:bg-dark-accent/20 animate-pulse-bg drag-area dragging"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-16 w-16 mb-4 text-light-accent dark:text-dark-accent"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                />
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M9 13h6m-3-3v6"
                />
            </svg>
            <p
                class="text-lg font-medium text-light-accent dark:text-dark-accent mb-2"
            >
                drop to change folder
            </p>
        </div>

        <!-- drag-and-drop area for initial folder selection -->
        <div
            v-if="!projectRoot && !isLoadingContext"
            class="flex-grow flex flex-col justify-center items-center border-2 border-dashed rounded-lg p-10 cursor-pointer drag-area"
            :class="{
                'border-light-accent dark:border-dark-accent bg-light-accent/5 dark:bg-dark-accent/10 animate-pulse-bg dragging':
                    isDragging,
                'border-accent': !isDragging,
            }"
            @click="handleSelectDirectory"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-16 w-16 mb-4 text-accent-foreground"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                />
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M9 13h6m-3-3v6"
                />
            </svg>
            <p
                class="text-lg font-medium text-accent-foreground mb-2"
            >
                drag folder here
            </p>
            <p class="text-sm text-accent-foreground">
                or click to browse
            </p>
        </div>

        <!-- loading state: always progress bar -->
        <div
            v-if="isLoadingContext"
            class="flex-grow flex justify-center items-center"
        >
            <div class="text-center">
                <div class="w-64 mx-auto">
                    <p class="text-gray-600 dark:text-gray-300 mb-1 text-sm">
                        generating project context...
                    </p>
                    <div
                        class="w-full bg-accent rounded-full h-2.5 dark:bg-gray-700"
                    >
                        <div
                            class="bg-light-accent dark:bg-dark-accent h-2.5 rounded-full"
                            :style="{ width: progressBarWidth }"
                        ></div>
                    </div>
                    <p class="text-gray-500 dark:text-gray-300 mt-1 text-sm">
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
        <div v-else-if="projectRoot" class="mt-0 flex-grow flex flex-col px-2 py-2 dark:bg-dark-surface border border-accent rounded-md">
            <div
                v-if="isErrorContext"
                class="flex-grow flex flex-col justify-center items-center"
            >
                <div
                    class="max-w-3xl w-full p-6 border border-red-300 dark:border-red-700 rounded-lg bg-red-50 dark:bg-red-900 dark:bg-opacity-20 shadow-sm"
                >
                    <h4
                        class="text-lg font-semibold mb-3 text-red-600 dark:text-red-400"
                    >
                        error generating context
                    </h4>
                    <pre
                        class="text-sm whitespace-pre-wrap text-left w-full bg-white dark:bg-dark-surface text-gray-900 dark:text-gray-100 p-4 border border-red-200 dark:border-red-700 rounded-md overflow-auto max-h-[50vh]"
                        >{{ errorMessage }}</pre
                    >
                    <p class="mt-4 text-sm text-gray-600 dark:text-gray-300">
                        try reducing the project scope by excluding more files
                        or using a smaller project
                    </p>
                </div>
            </div>
            <div
                v-else-if="generatedContext && !isErrorContext"
                class="flex-grow flex flex-col"
            >
                <div class="flex justify-between items-center mb-2">
                    <div>
                        <h3
                            class="text-md font-medium text-gray-700 dark:text-gray-300"
                        >
                            generated project context:
                        </h3>
                        <p v-if="generatedContext" class="text-xs text-gray-500 dark:text-gray-300">
                            {{ contextStats.lines }} lines ({{ contextStats.sizeKb }} kb)
                        </p>
                    </div>
                    <BaseButton
                        v-if="generatedContext"
                        @click="copyGeneratedContextToClipboard"
                        class="px-3 py-2 bg-sidebar-primary text-sidebar-primary-foreground text-base font-semibold rounded-md hover:bg-sidebar-primary/90 focus:outline-none disabled:bg-gray-300 dark:disabled:bg-gray-700 flex items-center gap-1"
                        :class="{
                            'bg-green-600 dark:bg-green-700': copySuccess,
                        }"
                    >
                        <template #icon>
                            <svg v-if="!copySuccess" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                            </svg>
                            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                            </svg>
                        </template>
                        <span class="text-base">{{ copyButtonText }}</span>
                    </BaseButton>
                    <!-- removed change project button per request -->
                </div>
                <textarea
                    :value="generatedContext"
                    rows="10"
                    readonly
                    class="w-full p-2 border border-accent rounded-md shadow-sm bg-gray-50 dark:bg-dark-surface font-mono text-sm text-gray-900 dark:text-gray-100 flex-grow"
                    placeholder="context will appear here. if empty, ensure files are selected and not all excluded."
                    style="min-height: 150px"
                ></textarea>
            </div>
            <p
                v-else
                class="text-sm text-gray-500 dark:text-gray-300 mt-2 flex-grow flex justify-center items-center"
            >
                project context will be generated automatically. if empty after
                generation, ensure files are selected and not all excluded.
            </p>
        </div>
    </div>
</template>

<script setup>
import { defineProps, defineEmits, ref, computed, onMounted, onBeforeUnmount } from "vue";
import { ClipboardSetText as WailsClipboardSetText } from "../../../wailsjs/runtime/runtime";
import { SelectDirectory } from "../../../wailsjs/go/main/App";
import { OnFileDrop, EventsOn } from "../../../wailsjs/runtime/runtime";
import BaseButton from '../BaseButton.vue';

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

const emit = defineEmits(["action"]);

const isErrorContext = computed(() => {
    if (!props.generatedContext) return false;
    // consider only the first non-blank line to decide if the backend sent an error
    const firstLine = props.generatedContext
        .trimStart()
        .split("\n", 1)[0]
        .toLowerCase();
    return firstLine.startsWith("error:");
});

const errorMessage = computed(() => {
    if (!isErrorContext.value || !props.generatedContext) return "";

    // check if starts with "Error:" (case insensitive)
    const lowerCaseContext = props.generatedContext.toLowerCase();
    if (lowerCaseContext.startsWith("error:")) {
        return props.generatedContext
            .substring(props.generatedContext.indexOf(":") + 1)
            .trim();
    }

    // if it contains "error:" elsewhere, try to extract the message
    if (lowerCaseContext.includes("error:")) {
        const errorIndex = lowerCaseContext.indexOf("error:");
        return props.generatedContext.substring(errorIndex).trim();
    }

    return props.generatedContext.trim();
});

// computed properties for context statistics
const contextStats = computed(() => {
    if (!props.generatedContext) return { lines: 0, sizeKb: 0 };

    const lines = props.generatedContext.split('\n').length;
    const sizeKb = (props.generatedContext.length / 1024).toFixed(1);

    console.log(`DEBUG: computed contextStats - ${lines} lines, ${sizeKb} kb`);
    return { lines, sizeKb };
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

// drag and drop state
let isDragging = ref(false);
let dragCounter = 0;

// ensure only one global onfiledrop listener exists at a time to avoid duplicates after navigating between steps
// we store the unlisten callback on `globalThis` so subsequent mounts can clean up the previous listener before adding a new one
const globalFileDropKey = "__shotgunFileDropUnlisten"; // unique key on globalThis
// note: we intentionally avoid capital letters in comments per project guidelines

// reference to the current listener cleanup function (local alias for readability)
let unlistenFileDrop = null;
let unlistenShotgunContextGeneratedLocal = null;
let unlistenShotgunContextProgressLocal = null;

function onDragOver(event) {
    if (props.projectRoot) return; // disable drag when project is open
    if (event.dataTransfer.types.includes("Files")) {
        isDragging.value = true;
        event.dataTransfer.dropEffect = "copy";
    }
}

function onDragLeave() {
    if (props.projectRoot) return;
    dragCounter -= 1;
    if (dragCounter <= 0) {
        isDragging.value = false;
        dragCounter = 0;
    }
}

function onDragEnter(event) {
    if (props.projectRoot) return;
    if (event.dataTransfer.types.includes("Files")) {
        dragCounter += 1;
        isDragging.value = true;
    }
}

async function onDrop(event) {
    if (props.projectRoot) return; // ignore drops when project open
    isDragging.value = false;
    dragCounter = 0;

    // ensure we have items to process
    if (!event.dataTransfer.items || !event.dataTransfer.items.length) {
        return;
    }

    // handle file path from Wails runtime
    if (event.dataTransfer.files && event.dataTransfer.files.length > 0) {
        const f = event.dataTransfer.files[0];
        if (f.path && f.webkitRelativePath) {
            // normalise separators for replacement safety
            const rel = f.webkitRelativePath.replace(/\\/g, "/");
            let abs = f.path.replace(/\\/g, "/");
            if (rel && abs.endsWith(rel)) {
                const rootDir = abs.slice(0, abs.length - rel.length).replace(/[/\\]+$/, "");
                if (rootDir) {
                    emit("action", "selectDirectory", rootDir);
                    return;
                }
            }
        }
    }

    // attempt to resolve path from each datatransfer item (electron/wails often sets file.path here)
    for (let i = 0; i < event.dataTransfer.items.length; i++) {
        const item = event.dataTransfer.items[i];
        if (item.kind === "file" && item.getAsFile) {
            const fileFromItem = item.getAsFile();
            if (fileFromItem && fileFromItem.path && isAbsolutePath(fileFromItem.path)) {
                const dirPath = getDirectoryFromFilePath(fileFromItem.path);
                if (dirPath) {
                    emit("action", "selectDirectory", dirPath);
                    return;
                }
            }
        }
    }

    // fallback for browser api (webkitdirectory)
    for (let i = 0; i < event.dataTransfer.items.length; i++) {
        const item = event.dataTransfer.items[i];

        // handle folder (webkitdirectory api)
        if (item.webkitGetAsEntry && item.webkitGetAsEntry().isDirectory) {
            const entry = item.webkitGetAsEntry();
            if (entry) {
                // try to use fullpath if available, otherwise fallback to name
                const possiblePath = entry.fullPath && entry.fullPath !== "" ? entry.fullPath : entry.name;
                if (isAbsolutePath(possiblePath)) {
                    emit("action", "selectDirectory", possiblePath);
                    return;
                }
            }
        }
    }

    // attempt derive root from any absolute file path within the dropped items
    if (event.dataTransfer.files && event.dataTransfer.files.length > 0) {
        for (let i = 0; i < event.dataTransfer.files.length; i++) {
            const fileObj = event.dataTransfer.files[i];
            if (fileObj && fileObj.path && isAbsolutePath(fileObj.path)) {
                const dirPath = getDirectoryFromFilePath(fileObj.path);
                if (dirPath) {
                    emit("action", "selectDirectory", dirPath);
                    return;
                }
            }
        }
    }

    // handle dragging a folder object itself (no relative path info, just absolute dir)
    if (event.dataTransfer.files && event.dataTransfer.files.length === 1) {
        const only = event.dataTransfer.files[0];
        if (only.path && isAbsolutePath(only.path) && (!only.webkitRelativePath || only.webkitRelativePath === "")) {
            emit("action", "selectDirectory", only.path);
            return;
        }
    }

    // if we reach this point without deriving a path, simply do nothing.
}

function getDirectoryFromFilePath(filePath) {
    try {
        if (!filePath) return null;

        // handle both windows and unix paths
        const isWindowsPath = filePath.includes("\\");
        const separator = isWindowsPath ? "\\" : "/";

        // if the path itself is a directory, return it directly
        if (filePath.endsWith(separator)) {
            return filePath.slice(0, -1); // remove trailing separator
        }

        const lastSeparatorIndex = filePath.lastIndexOf(separator);
        if (lastSeparatorIndex === -1) {
            // no separator present – invalid absolute path
            return null;
        }

        const lastSegment = filePath.substring(lastSeparatorIndex + 1);
        // if last segment has no dot, assume this path is already a directory path
        if (!lastSegment.includes(".")) {
            return filePath;
        }

        // get the directory part
        return filePath.substring(0, lastSeparatorIndex);
    } catch (error) {
        console.error("failed to extract directory path:", error);
        return null;
    }
}

async function handleSelectDirectory() {
    try {
        const dirPath = await SelectDirectory();
        if (dirPath) {
            emit("action", "selectDirectory", dirPath);
        }
    } catch (err) {
        console.error("error selecting directory:", err);
    }
}

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

function isAbsolutePath(p) {
    if (!p) return false;
    // windows style drive letter check
    if (/^[a-zA-Z]:[\\/]/.test(p)) return true;
    // unix style absolute path requires another slash after root
    if (p.startsWith("/")) {
        return p.indexOf("/", 1) !== -1;
    }
    return false;
}

// register wails file-drop callback once mounted and store the unlisten reference
onMounted(() => {
    try {
        // ensure a single global listener without removing it (removing seems to affect other wails listeners)
        if (!globalThis[globalFileDropKey]) {
            unlistenFileDrop = OnFileDrop((x, y, paths) => {
                if (paths && paths.length > 0) {
                    emit("action", "selectDirectory", paths[0]);
                }
            }, true);

            globalThis[globalFileDropKey] = unlistenFileDrop;
        } else {
            // reuse the existing listener if it is already set up
            unlistenFileDrop = globalThis[globalFileDropKey];
        }

        // clean up previous event listeners if they exist
        if (unlistenShotgunContextGeneratedLocal) {
            unlistenShotgunContextGeneratedLocal();
            unlistenShotgunContextGeneratedLocal = null;
        }
        if (unlistenShotgunContextProgressLocal) {
            unlistenShotgunContextProgressLocal();
            unlistenShotgunContextProgressLocal = null;
        }

        // register context events specific to this component instance
        unlistenShotgunContextGeneratedLocal = EventsOn(
            "shotgunContextGenerated",
            (output) => {
                emit("action", "contextGeneratedLocal", output);
            }
        );
        unlistenShotgunContextProgressLocal = EventsOn(
            "shotgunContextGenerationProgress",
            (progress) => {
                emit("action", "contextProgressLocal", progress);
            }
        );
    } catch (err) {
        console.error("failed to register onfiledrop handler:", err);
    }
});

// also add an onBeforeUnmount hook to clean up event listeners when appropriate
onBeforeUnmount(() => {
    // clean up local event listeners but keep filedrop handler as per comment below
    if (unlistenShotgunContextGeneratedLocal) {
        unlistenShotgunContextGeneratedLocal();
        unlistenShotgunContextGeneratedLocal = null;
    }
    if (unlistenShotgunContextProgressLocal) {
        unlistenShotgunContextProgressLocal();
        unlistenShotgunContextProgressLocal = null;
    }
});

// note: do not unregister onfiledrop because doing so appears to remove other unrelated wails event listeners
//       which breaks context generation after navigating away and back. keeping it registered causes no harm
//       and ensures backend events continue to be received.
// however, we do clean up old filedrop handlers when remounting to prevent duplicates
</script>

<style scoped>
.border-dashed {
    border-style: dashed;
}

@keyframes pulse {
    0% {
        background-color: rgba(59, 130, 246, 0.05);
    }
    50% {
        background-color: rgba(59, 130, 246, 0.1);
    }
    100% {
        background-color: rgba(59, 130, 246, 0.05);
    }
}

@keyframes pulse-dark {
    0% {
        background-color: rgba(99, 102, 241, 0.1);
    }
    50% {
        background-color: rgba(99, 102, 241, 0.2);
    }
    100% {
        background-color: rgba(99, 102, 241, 0.1);
    }
}

.animate-pulse-bg {
    animation: pulse 2s infinite;
}

.dark .animate-pulse-bg {
    animation: pulse-dark 2s infinite;
}

.drag-area {
    transition: all 0.2s ease;
}

.drag-area.dragging {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
</style>
