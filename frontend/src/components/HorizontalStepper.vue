<template>
    <nav
        class="bg-light-surface dark:bg-[#141414] dark:shadow-gray-800 top-0 z-10 center relative"
    >
        <div class="container mx-auto px-3 py-4">
            <ol class="flex items-center justify-between">
                <li
                    v-for="(step, index) in steps"
                    :key="step.id"
                    class="flex-1 group"
                    :class="{ 'flex items-center': index < steps.length - 1 }"
                >
                    <div class="flex items-center w-full justify-center">
                        <button
                            @click.prevent="
                                canNavigateToStep(step.id)
                                    ? $emit('navigate', step.id)
                                    : null
                            "
                            :class="[
                                'flex items-center justify-center text-sm sm:text-sm font-medium text-center px-4 py-2 rounded-lg w-full max-w-[180px] mx-auto border-2 border-transparent',
                                canNavigateToStep(step.id)
                                    ? 'cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                                    : 'cursor-not-allowed opacity-60',
                            ]"
                            :disabled="!canNavigateToStep(step.id)"
                            :title="step.description"
                        >
                            <div
                                :class="[
                                    'flex items-center justify-center w-7 h-7 sm:w-8 sm:h-8 rounded-md border-2 mr-2 shrink-0',
                                    currentStep === step.id
                                        ? 'border-light-accent bg-indigo-100 dark:bg-indigo-900 dark:bg-opacity-30 text-indigo-700 dark:text-indigo-300'
                                        : step.completed
                                        ? 'border-green-600 bg-green-100 dark:bg-green-900 dark:bg-opacity-30 text-green-700 dark:text-green-300'
                                        : 'border-gray-400 dark:border-gray-500 bg-gray-50 dark:bg-gray-700 text-gray-500 dark:text-gray-400 group-hover:border-gray-500 dark:group-hover:border-gray-400',
                                ]"
                            >
                                <span
                                    v-if="
                                        !(
                                            step.completed &&
                                            currentStep !== step.id
                                        )
                                    "
                                    >{{ step.id }}</span
                                >
                                <svg
                                    v-else
                                    class="w-3 h-3 sm:w-4 sm:h-4"
                                    fill="currentColor"
                                    viewBox="0 0 20 20"
                                >
                                    <path
                                        fill-rule="evenodd"
                                        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                        clip-rule="evenodd"
                                    ></path>
                                </svg>
                            </div>
                            <span
                                :class="[
                                    'whitespace-nowrap text-sm font-medium',
                                    currentStep === step.id
                                        ? 'text-light-accent dark:text-dark-accent'
                                        : step.completed
                                        ? 'text-green-600 dark:text-green-400'
                                        : 'text-gray-500 dark:text-gray-400 group-hover:text-gray-700 dark:group-hover:text-gray-300',
                                ]"
                            >
                                {{ step.title }}
                            </span>
                        </button>
                    </div>
                    <!-- connector line -->
                    <div
                        v-if="index < steps.length - 1"
                        class="flex-auto border-t-2 transition-all duration-300 ease-in-out mt-4 mx-2"
                        :class="
                            step.completed
                                ? 'border-green-500 dark:border-green-700'
                                : 'border-gray-300 dark:border-gray-600'
                        "
                    ></div>
                </li>
            </ol>
        </div>
    </nav>
</template>
<script setup>
import { defineProps, defineEmits } from "vue";

const props = defineProps({
    currentStep: { type: Number, required: true },
    steps: { type: Array, required: true }, // array of { id: number, title: string, completed: boolean }
});

const emit = defineEmits(["navigate"]);

function canNavigateToStep(stepId) {
    // allow navigating to the current step
    if (stepId === props.currentStep) return true;

    const targetStep = props.steps.find((s) => s.id === stepId);

    // allow navigating to steps that are either completed or have been visited before
    if (targetStep && (targetStep.completed || targetStep.visited)) return true;

    // otherwise allow navigation only to the first uncompleted step (normal forward flow)
    const firstUncompletedStepId = props.steps.find((s) => !s.completed)?.id;

    return (
        stepId === firstUncompletedStepId ||
        (firstUncompletedStepId === undefined && !!targetStep)
    ); // allow any step if all are completed
}
</script>
