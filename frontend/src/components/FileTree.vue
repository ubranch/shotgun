<template>
    <ul class="file-tree overflow-y-hidden">
        <li
            v-for="node in nodes"
            :key="node.path"
            :class="{ 'excluded-node': node.excluded }"
        >
            <div
                class="node-item p-1 border-b border-gray-200 dark:border-gray-700"
                :style="{ 'padding-left': depth * 20 + 'px' }"
                style="position: relative; cursor: pointer"
                @click="handleAreaClick($event, node)"
            >
                <span class="arrow-indicator pl-1">
                    <span v-if="node.isDir && !node.expanded" @click.stop="toggleExpand(node)">
                        <div class="codicon codicon-chevron-down codicon-bold codicon-custom"></div>
                    </span>
                    <span v-else-if="node.isDir && node.expanded" @click.stop="toggleExpand(node)">
                        <div class="codicon codicon-chevron-up codicon-bold codicon-custom"></div>
                    </span>
                    <span v-else class="placeholder-arrow"></span>
                </span>

                <span class="node-content-wrapper pl-3">
                    <span
                        v-if="node.isDir"
                        @click.stop="toggleExpand(node)"
                        class="toggler"
                    >
                        <!-- folder icon (closed) -->
                        <i class="codicon codicon-folder text-decoration-none no-underline"
                            v-if="!node.expanded"
                        />
                        <!-- folder icon (open) -->
                        <i class="codicon codicon-folder-opened no-underline" v-else/>
                    </span>
                    <!-- file icon -->
                    <span v-else class="file-icon">
                        <i class="codicon codicon-file text-decoration-none no-underline"></i>
                    </span>
                    <span
                        @click.stop="node.isDir ? toggleExpand(node) : handleCheckboxChange(node)"
                        :class="{ 'folder-name': node.isDir }"
                        class="text-sm name-label"
                    >
                        {{ node.name }}
                    </span>
                </span>

                <span class="checkbox-wrapper" @click.stop>
                    <input
                        type="checkbox"
                        :checked="!node.excluded"
                        @change="handleCheckboxChange(node)"
                        class="exclude-checkbox"
                    />
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
    // emit an event with the node to toggle its exclusion status in the parent component
    emit("toggle-exclude", node);
}

function handleAreaClick(event, node) {
    // decide behavior based on node type and click target
    const target = event.target;

    // ignore clicks on the checkbox area (they have their own events)
    if (target.closest('.checkbox-wrapper') || target.classList.contains('exclude-checkbox')) {
        return;
    }

    if (node.isDir) {
        // if the click wasn't on the arrow icon, folder name, or explicit toggler, treat the whole row as an expander
        if (
            !target.closest('.arrow-indicator') &&
            !target.classList.contains('folder-name') &&
            !target.classList.contains('toggler')
        ) {
            toggleExpand(node);
            return;
        }
        // the other folder-specific elements already have their own @click handlers
    } else {
        // for files, whitespace toggles inclusion/exclusion
        handleCheckboxChange(node);
    }
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
.checkbox-wrapper {
    margin-left: auto; /* push to the end of the line */
    display: flex;
    align-items: center;
}
.exclude-checkbox {
    cursor: pointer;
    width: 20px;
    height: 20px;
}
.excluded-node > .node-item > span:not(.toggler, .file-icon, .arrow-indicator, .codicon) {
    color: #999;
}
.exclude-checkbox:disabled {
    cursor: not-allowed;
}
.name-label {
    margin-left: 8px;
}
.arrow-indicator {
    display: flex;
    align-items: center;
    width: 22px;
}
.placeholder-arrow {
    width: 22px; /* same as the real arrow, to maintain alignment */
}
.node-content-wrapper {
    display: flex;
    align-items: center;
    flex-grow: 1; /* take up available space */
    min-width: 0; /* allow text to be truncated */
    overflow: hidden; /* for text truncation if needed */
}
.codicon-bold {
    font-weight: bold !important; /* or a numeric value like 700 */
}

.codicon-custom {
    align-items: center;
    display: flex !important;
    flex-shrink: 0;
    font-size: 22px;
    justify-content: center;
    padding-right: 6px;
    text-align: right;
    transform: translateX(3px);
    width: 22px;
    height: 100%;
    text-decoration: none!important;
}
</style>
