<template>
    <div>
        <slot></slot>
    </div>
</template>

<script setup>
import { ref, provide, onMounted, watch } from "vue";

// initialize theme state
const isDark = ref(true);

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
        document.documentElement.classList.remove("light");
    } else {
        document.documentElement.classList.add("light");
        document.documentElement.classList.remove("dark");
    }
};

// initialize theme on mount
onMounted(() => {
    // check for saved theme preference or use system preference
    const savedTheme = localStorage.getItem("theme");
    if (savedTheme) {
        isDark.value = savedTheme === "dark";
    } else {
        // use system preference as fallback
        isDark.value = window.matchMedia("(prefers-color-scheme: dark)").matches;
    }

    // apply the theme
    applyTheme();

    // listen for system theme changes
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    mediaQuery.addEventListener("change", (e) => {
        // only update if no saved preference exists
        if (!localStorage.getItem("theme")) {
            isDark.value = e.matches;
            applyTheme();
        }
    });
});
</script>
