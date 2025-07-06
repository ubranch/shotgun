<template>
    <div>
        <slot></slot>
    </div>
</template>

<script setup>
import { ref, provide, onMounted, watch } from "vue";

// initialize to true for dark mode
const isDark = ref(true);

// provide theme context to child components
provide("isDark", isDark);

// provide toggle function to child components
// modified to be a no-op that keeps dark mode
const toggleTheme = () => {
    // always maintain dark mode
    isDark.value = true;
    localStorage.setItem("theme", "dark");
    applyTheme();
};

provide("toggleTheme", toggleTheme);

// apply theme to document
const applyTheme = () => {
    // always add dark class
    document.documentElement.classList.add("dark");
};

// initialize theme - always use dark mode
onMounted(() => {
    // force dark mode regardless of saved preference
    isDark.value = true;
    localStorage.setItem("theme", "dark");
    applyTheme();

    // no need to listen for system theme changes as we always use dark mode
});
</script>
