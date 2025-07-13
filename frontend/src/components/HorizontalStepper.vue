<template>
    <div class="bg-card top-0 z-10 center relative">
        <div class="flex items-center justify-between px-4 py-2">
            <div class="flex items-center">
                <ol class="flex space-x-2 items-center">
                    <li
                        v-for="step in steps"
                        :key="step.id"
                        class="flex items-center"
                    >
                        <button
                            @click="navigateToStep(step.id)"
                            :disabled="!canNavigateTo(step.id)"
                            :class="[
                                'px-3 py-1 rounded-md text-sm transition-colors',
                                currentStep === step.id
                                    ? 'bg-primary text-primary-foreground'
                                    : step.visited
                                      ? 'bg-accent text-accent-foreground hover:bg-accent/80'
                                      : 'bg-muted text-muted-foreground',
                            ]"
                        >
                            {{ step.id }}. {{ step.title }}
                        </button>
                        <div
                            v-if="step.id !== steps.length"
                            class="mx-2 text-muted-foreground"
                        >
                            &nbsp;&nbsp;→
                        </div>
                    </li>
                    <li class="flex items-center">
                        <div class="mx-2 text-muted-foreground">or&nbsp;&nbsp;</div>
                        <BaseButton
                            @click="reset"
                            variant="danger"
                            size="sm"
                        >
                            reset
                        </BaseButton>
                    </li>
                </ol>
            </div>
        </div>
    </div>
</template>
<script setup>
import { defineProps, defineEmits } from "vue";
import BaseButton from "./BaseButton.vue";

const props = defineProps({
    currentStep: {
        type: Number,
        required: true,
    },
    steps: {
        type: Array,
        required: true,
    },
});

const emit = defineEmits(["navigate", "reset"]);

function navigateToStep(stepId) {
    if (canNavigateTo(stepId)) {
        emit("navigate", stepId);
    }
}

function reset() {
    emit("reset");
}

function canNavigateTo(stepId) {
    // allow navigation to the current step or any step that has been visited
    return (
        stepId === props.currentStep ||
        props.steps.find((step) => step.id === stepId)?.visited ||
        props.steps.find((step) => step.id === stepId - 1)?.completed
    );
}
</script>
