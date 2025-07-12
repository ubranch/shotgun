<template>
    <div
        v-if="isVisible"
        class="fixed inset-0 bg-background/80 overflow-y-auto h-full w-full z-50 flex justify-center items-center"
        @click.self="handleCancel"
    >
        <div
            class="relative mx-auto p-5 border w-full max-w-2xl shadow-lg rounded-md bg-card border-border"
        >
            <div class="mt-3 text-center">
                <h3
                    class="text-lg leading-6 font-medium text-card-foreground"
                >
                    {{ title }}
                </h3>
                <div class="mt-2 px-7 py-3">
                    <textarea
                        v-model="editableRules"
                        rows="15"
                        spellcheck="false"
                        class="w-full p-2 border border-border rounded-md shadow-sm focus:ring-primary focus:border-primary text-sm font-mono bg-background text-foreground"
                        :placeholder="`enter ${ruleType} patterns here...`"
                    ></textarea>
                    <p
                        class="text-sm text-muted-foreground mt-1 text-left"
                    >
                        {{ descriptionText }}
                    </p>
                </div>
                <div class="items-center px-4 py-3">
                    <BaseButton
                        @click="handleSave"
                        variant="primary"
                        class="px-4 py-2 mr-2"
                    >
                        save
                    </BaseButton>
                    <BaseButton
                        @click="handleCancel"
                        class="px-4 py-2"
                    >
                        cancel
                    </BaseButton>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, defineProps, defineEmits, computed } from "vue";
import BaseButton from './BaseButton.vue';

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
