<template>
    <div
        v-if="isVisible"
        class="fixed inset-0 bg-gray-600 bg-opacity-50 dark:bg-dark-surface dark:bg-opacity-80 overflow-y-auto h-full w-full z-50 flex justify-center items-center"
        @click.self="handleCancel"
    >
        <div
            class="relative mx-auto p-5 border w-full max-w-2xl shadow-lg rounded-md bg-white dark:bg-dark-surface border-gray-200 dark:border-gray-700"
        >
            <div class="mt-3 text-center">
                <h3
                    class="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100"
                >
                    {{ title }}
                </h3>
                <div class="mt-2 px-7 py-3">
                    <textarea
                        v-model="editableRules"
                        rows="15"
                        class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm font-mono bg-gray-50 dark:bg-dark-surface text-gray-900 dark:text-gray-100"
                        placeholder="enter custom ignore patterns, one per line (e.g., *.log, node_modules/)"
                    ></textarea>
                    <p
                        class="text-sm text-gray-500 dark:text-gray-400 mt-1 text-left"
                    >
                        {{ descriptionText }}
                    </p>
                </div>
                <div class="items-center px-4 py-3">
                    <button
                        @click="handleSave"
                        class="px-4 py-2 bg-blue-500 dark:bg-blue-600 text-white text-base font-medium rounded-md w-auto hover:bg-blue-700 dark:hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 mr-2"
                    >
                        save
                    </button>
                    <button
                        @click="handleCancel"
                        class="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 text-base font-medium rounded-md w-auto hover:bg-gray-300 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-400"
                    >
                        cancel
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, defineProps, defineEmits, computed } from "vue";

const props = defineProps({
    isVisible: {
        type: Boolean,
        required: true,
    },
    initialRules: {
        type: String,
        default: "",
    },
    title: {
        type: String,
        default: "edit custom rules",
    },
    ruleType: {
        type: String,
        required: true,
        validator: (value) => ["ignore", "prompt"].includes(value),
    },
});

const emit = defineEmits(["save", "cancel"]);

const editableRules = ref("");

const descriptionText = computed(() => {
    if (props.ruleType === "prompt") {
        return "these rules provide specific instructions or pre-defined text for the ai. they will be included in the final prompt.";
    }
    // default to the description for ignore rules
    return 'these rules use .gitignore pattern syntax. they are applied globally when "use custom rules" is checked.';
});

watch(
    () => props.initialRules,
    (newVal) => {
        editableRules.value = newVal;
    },
    { immediate: true }
);

watch(
    () => props.isVisible,
    (newVal) => {
        if (newVal) {
            // when modal becomes visible, ensure textarea reflects the latest initialrules
            editableRules.value = props.initialRules;
        }
    }
);

function handleSave() {
    emit("save", editableRules.value);
}

function handleCancel() {
    emit("cancel");
}
</script>

<style scoped>
/* basic styling for modal, can be enhanced with tailwind further if needed */
</style>
