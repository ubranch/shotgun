<template >
    <ul class="file-tree overflow-y-hidden">
        <li
            v-for="node in nodes"
            :key="node.path"
            :class="{ 'excluded-node': node.excluded }"
        >
            <div
                class="node-item"
                :style="{ 'padding-left': depth * 20 + 'px' }"
            >
            <input
                    type="checkbox"
                    :checked="!node.excluded"
                    @change="handleCheckboxChange(node)"
                    class="exclude-checkbox"
                />
                <span
                    v-if="node.isDir"
                    @click="toggleExpand(node)"
                    class="toggler"
                >
                    <!-- folder icon (closed) -->
                    <svg
                        v-if="!node.expanded"
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-4 w-4 text-yellow-600 dark:text-yellow-500"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                    >
                        <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
                    </svg>
                    <!-- folder icon (open) -->
                    <svg
                        v-else
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-4 w-4 text-yellow-600 dark:text-yellow-500"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                    >
                        <path fill-rule="evenodd" d="M2 6a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1h-14v-1zm0 3v4a2 2 0 002 2h12a2 2 0 002-2v-4h-14z" clip-rule="evenodd" />
                    </svg>
                </span>
                <!-- file icon -->
                <span v-else class="file-icon">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-4 w-4 text-gray-600 dark:text-gray-400"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                    >
                        <path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z" clip-rule="evenodd" />
                    </svg>
                </span>

                <span
                    @click="node.isDir ? toggleExpand(node) : null"
                    :class="{ 'folder-name': node.isDir }"
                    class="text-sm"
                >
                    {{ node.name }}
                </span>
            </div>
            <FileTree
                v-if="node.isDir && node.expanded && node.children"
                :nodes="node.children"
                :project-root="projectRoot"
                :depth="depth + 1"
                @toggle-exclude="emitToggleExclude"
            />
        </li>
    </ul>
</template>

<script setup>
import { defineProps, defineEmits } from "vue";

const props = defineProps({
    nodes: Array,
    projectRoot: String,
    depth: {
        type: Number,
        default: 0,
    },
    parentExcluded: {
        // whether an ancestor is excluded
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["toggle-exclude"]);

function toggleExpand(node) {
    if (node.isDir) {
        node.expanded = !node.expanded;
    }
}

function handleCheckboxChange(node) {
    // emit an event with the node to toggle its exclusion status in the parent (app.vue)
    emit("toggle-exclude", node);
}

function emitToggleExclude(node) {
    emit("toggle-exclude", node); // bubble up the event
}

// a node is effectively excluded if one of its parents is.
// this is mainly for ui state (e.g., disabling checkbox), backend handles true exclusion.
function isEffectivelyExcludedByParent(node) {
    let current = node.parent;
    while (current) {
        if (current.excluded) return true;
        current = current.parent;
    }
    return false;
}
</script>

<style scoped>
.file-tree {
    list-style-type: none;
}

.node-item {
    display: flex;
    align-items: center;
    cursor: default;
    transition: background-color 0.15s ease;
}
.node-item:hover {
    background-color: rgba(0, 0, 0, 0.05);
}
.toggler {
    cursor: pointer;
    width: 20px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
}
.file-icon {
    width: 20px; /* space for files to align with folder icons */
    display: inline-flex;
    align-items: center;
    justify-content: center;
}
.folder-name {
    cursor: pointer; /* to indicate it's clickable for expanding */
    font-weight: bold;
}
.exclude-checkbox {
    margin-right: 5px;
    margin-left: 5px;
    cursor: pointer;
}
.excluded-node > .node-item > span:not(.toggler):not(.file-icon) {
    text-decoration: line-through;
    color: #999;
}
.exclude-checkbox:disabled {
    cursor: not-allowed;
}
</style>
