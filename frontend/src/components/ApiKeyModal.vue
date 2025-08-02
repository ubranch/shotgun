<template>
    <div v-if="show" class="fixed inset-0 bg-black/50 backdrop-blur-sm overflow-y-auto h-full w-full z-50 flex justify-center items-center" @click.self="close">
        <div class="relative mx-auto p-5 border w-full max-w-md shadow-lg rounded-md bg-card border-border">
            <div class="text-center">
                <h3 class="text-lg leading-6 font-medium text-card-foreground">set gemini api key</h3>
                <div class="mt-2 px-7 py-3">
                    <input
                        :type="showApiKey ? 'text' : 'password'"
                        v-model="apiKey"
                        placeholder="enter your gemini api key"
                        class="w-full p-2 border border-border rounded-md shadow-sm focus:ring-primary focus:border-primary text-sm font-mono bg-background text-foreground"
                    />
                    <label class="flex items-center space-x-1 text-sm mt-2 text-muted-foreground">
                        <input
                            type="checkbox"
                            v-model="showApiKey"
                            class="rounded border-gray-300 text-accent focus:ring-accent"
                        />
                        <span>show api key</span>
                    </label>
                </div>
                <div class="items-center px-4 py-3 flex justify-center space-x-4">
                    <BaseButton
                        @click="save"
                        class="px-4 py-2 bg-sidebar-primary text-sidebar-primary-foreground text-base font-semibold rounded-md hover:bg-sidebar-primary/90 focus:outline-none"
                    >
                        <span class="text-base">save</span>
                    </BaseButton>
                    <BaseButton
                        @click="close"
                        class="px-4 py-2"
                    >
                        <span class="text-base">close</span>
                    </BaseButton>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import BaseButton from './BaseButton.vue';

const props = defineProps({
    show: Boolean,
    initialApiKey: String,
});

const emit = defineEmits(['close', 'save']);

const apiKey = ref(props.initialApiKey);
const showApiKey = ref(false);

watch(() => props.initialApiKey, (newVal) => {
    apiKey.value = newVal;
});

function save() {
    emit('save', apiKey.value);
}

function close() {
    emit('close');
}
</script>
