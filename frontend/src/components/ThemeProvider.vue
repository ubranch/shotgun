<template>
    <div>
        <slot></slot>
    </div>
</template>

<script setup>
import { ref, provide, onMounted, watch } from "vue";

const isDark = ref(false);

// provide theme context to child components
provide("isDark", isDark);

// provide toggle function to child components
const toggleTheme = () => {
    isDark.value = !isDark.value;
    localStorage.setItem("theme", isDark.value ? "dark" : "light");
    applyTheme();
};

provide("toggleTheme", toggleTheme);

// apply theme to document
const applyTheme = () => {
    if (isDark.value) {
        document.documentElement.classList.add("dark");
    } else {
        document.documentElement.classList.remove("dark");
    }
};

// initialize theme from local storage or system preference
onMounted(() => {
    // check local storage first
    const savedTheme = localStorage.getItem("theme");

    if (savedTheme) {
        isDark.value = savedTheme === "dark";
    } else {
        // check system preference
        isDark.value = window.matchMedia(
            "(prefers-color-scheme: dark)"
        ).matches;
        localStorage.setItem("theme", isDark.value ? "dark" : "light");
    }

    applyTheme();

    // listen for system theme changes
    window
        .matchMedia("(prefers-color-scheme: dark)")
        .addEventListener("change", (e) => {
            if (!localStorage.getItem("theme")) {
                isDark.value = e.matches;
                applyTheme();
            }
        });
});
</script>
