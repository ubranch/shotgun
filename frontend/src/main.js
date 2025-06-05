import { createApp } from "vue";
import App from "./App.vue";
import "./assets/main.css"; // Basic styles and tailwind
import "./assets/custom.css"; // Custom styles - loaded last to override previous styles

createApp(App).mount("#app");
