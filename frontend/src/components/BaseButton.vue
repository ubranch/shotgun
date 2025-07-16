<template>
    <button
        data-slot="button"
        :class="[
            'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all duration-200',
            'disabled:pointer-events-none disabled:opacity-50',
            '[&_svg]:pointer-events-none [&_svg:not([class*=\'size-\'])]:size-4 [&_svg]:shrink-0',
            'outline-none focus-visible:ring-2 focus-visible:ring-offset-2',
            'border-2',
            'shadow-sm',
            variantClasses,
            sizeClasses,
        ]"
        :disabled="disabled"
        v-bind="$attrs"
    >
        <slot name="icon"></slot>
        <slot></slot>
    </button>
</template>

<script setup>
import { defineProps, computed } from "vue";

const props = defineProps({
    disabled: {
        type: Boolean,
        default: false,
    },
    variant: {
        type: String,
        default: "default",
        validator: (value) =>
            ["default", "primary", "success", "danger", "warning"].includes(
                value
            ),
    },
    size: {
        type: String,
        default: "md",
        validator: (value) => ["sm", "md", "lg"].includes(value),
    },
});

const variantClasses = computed(() => {
    switch (props.variant) {
        case "primary":
            return "border-primary bg-primary/10 text-primary hover:bg-primary/20 hover:border-primary/80 focus-visible:ring-primary";
        case "success":
            return "border-green-500 bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-800/50 hover:border-green-600 dark:hover:border-green-400 focus-visible:ring-green-500";
        case "danger":
            return "border-destructive bg-destructive/10 text-destructive hover:bg-destructive/20 hover:border-destructive/80 focus-visible:ring-destructive";
        case "warning":
            return "border-amber-500 bg-amber-50 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300 hover:bg-amber-100 dark:hover:bg-amber-800/50 hover:border-amber-600 dark:hover:border-amber-400 focus-visible:ring-amber-500";
        default:
            return "border-border bg-background hover:bg-accent hover:border-primary/50 text-foreground hover:text-primary focus-visible:ring-primary";
    }
});

const sizeClasses = computed(() => {
    switch (props.size) {
        case "sm":
            return "h-7 px-3 py-1 text-xs";
        case "lg":
            return "h-11 px-4  py-2.5 text-base";
        default:
            return "h-9 px-3 py-2 max-sm:p-0 aspect-square";
    }
});
</script>

<style scoped>
button {
    position: relative;
}

button:active::after {
    transform: scale(0, 0);
    opacity: 0.3;
    transition: 0s;
}
</style>
