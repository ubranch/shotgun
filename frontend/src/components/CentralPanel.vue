<template>
  <main class="flex-1 p-0 overflow-y-auto bg-white">
    <Step1CopyStructure v-if="currentStep === 1" @action="handleAction" :generated-context="shotgunPromptContext" :step1-context-generation-attempted="props.step1ContextGenerationAttempted" :project-root="props.projectRoot" />
    <Step2GenerateDiff v-if="currentStep === 2" @action="handleAction" ref="step2Ref" />
    <Step3ExecuteDiff v-if="currentStep === 3" @action="handleAction" ref="step3Ref" />
    <Step4ApplyPatch v-if="currentStep === 4" @action="handleAction" />
  </main>
</template>

<script setup>
import { defineProps, defineEmits, ref } from 'vue';
import Step1CopyStructure from './steps/Step1CopyStructure.vue';
import Step2GenerateDiff from './steps/Step2GenerateDiff.vue';
import Step3ExecuteDiff from './steps/Step3ExecuteDiff.vue';
import Step4ApplyPatch from './steps/Step4ApplyPatch.vue';

const props = defineProps({
  currentStep: { type: Number, required: true },
  shotgunPromptContext: { type: String, default: '' },
  step1ContextGenerationAttempted: { type: Boolean, default: false },
  projectRoot: { type: String, default: '' }
});

const emit = defineEmits(['stepAction']);

const step2Ref = ref(null);
const step3Ref = ref(null);

function handleAction(actionName, payload) {
  emit('stepAction', actionName, payload);
}

const updateStep2DiffOutput = (output) => {
  if (step2Ref.value && step2Ref.value.setDiffOutput) {
    step2Ref.value.setDiffOutput(output);
  }
};

const updateStep2ShotgunContext = (context) => {
  if (step2Ref.value && step2Ref.value.setShotgunContext) {
    step2Ref.value.setShotgunContext(context);
  }
};

const addLogToStep3Console = (message, type) => {
  if (step3Ref.value && step3Ref.value.addLog) {
    step3Ref.value.addLog(message, type);
  }
};

defineExpose({ updateStep2DiffOutput, addLogToStep3Console, updateStep2ShotgunContext });
</script> 