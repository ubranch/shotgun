import { createApp } from "vue";
import App from "./App.vue";
import "./assets/main.css";
import "./assets/custom.css";

// register once for backend events and forward them as custom window events
import { EventsOn } from "../wailsjs/runtime/runtime";

// track global event listeners to prevent duplicate registrations
let globalListeners = {
  shotgunContextGenerated: null,
  shotgunContextGenerationProgress: null
};

function registerGlobalShotgunListeners() {
  console.log("registering global shotgun listeners");

  // clean up any existing listeners first
  unregisterGlobalShotgunListeners();

  try {
    globalListeners.shotgunContextGenerated = EventsOn("shotgunContextGenerated", (output) => {
      window.dispatchEvent(
        new CustomEvent("shotgun-context-generated", { detail: output })
      );
    });

    console.log("listener for shotgunContextGenerated registered");

    globalListeners.shotgunContextGenerationProgress = EventsOn("shotgunContextGenerationProgress", (progress) => {
      window.dispatchEvent(
        new CustomEvent("shotgun-context-progress", { detail: progress })
      );
    });
    console.log("listener for shotgunContextGenerationProgress registered");
  } catch (err) {
    console.error("failed to register global shotgun context listeners:", err);
  }
}

function unregisterGlobalShotgunListeners() {
  // clean up existing listeners
  try {
    if (globalListeners.shotgunContextGenerated) {
      globalListeners.shotgunContextGenerated();
      globalListeners.shotgunContextGenerated = null;
      console.log("unregistered shotgunContextGenerated listener");
    }

    if (globalListeners.shotgunContextGenerationProgress) {
      globalListeners.shotgunContextGenerationProgress();
      globalListeners.shotgunContextGenerationProgress = null;
      console.log("unregistered shotgunContextGenerationProgress listener");
    }
  } catch (err) {
    console.error("error unregistering global shotgun listeners:", err);
  }
}

// disable browser zoom
const disableZoom = () => {
  // prevent ctrl + wheel zoom
  document.addEventListener('wheel', (event) => {
    if (event.ctrlKey) {
      event.preventDefault();
    }
  }, { passive: false });

  // prevent ctrl + +/- zoom
  document.addEventListener('keydown', (event) => {
    if (event.ctrlKey && (event.key === '+' || event.key === '-' || event.key === '=')) {
      event.preventDefault();
    }
  }, { passive: false });
};

// call the function when dom is loaded
if (document.readyState === 'complete') {
  disableZoom();
} else {
  document.addEventListener('DOMContentLoaded', disableZoom);
}

const app = createApp(App);
app.mount("#app");

// register listeners after vue app is mounted (runtime should be ready)
registerGlobalShotgunListeners();

// clean up listeners when window is closed/refreshed
window.addEventListener('beforeunload', () => {
  unregisterGlobalShotgunListeners();
});
