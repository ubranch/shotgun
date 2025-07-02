<template>
    <ThemeProvider>
        <MainLayout />
    </ThemeProvider>
</template>

<script setup>
import MainLayout from "./components/MainLayout.vue";
import ThemeProvider from "./components/ThemeProvider.vue";
</script>

<style scoped>
/* styles from the old app.vue can be moved to specific components or global styles if needed */
</style>
