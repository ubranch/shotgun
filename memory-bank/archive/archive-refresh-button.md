# archive: add refresh button before prompt modes

## task details

**task**: add a refresh button before the prompt modes section
**complexity level**: level 2 - simple enhancement
**files modified**: frontend/src/components/steps/Step2ComposePrompt.vue
**completion date**: june 2025

## requirements

add a "refresh" button with a refresh icon before the "prompt" button in the modes section. when pressed, the button should invoke the regeneration of the prompt. the button should be positioned to the left of the existing mode buttons.

## implementation details

### state management
```js
// refresh button state
const refreshing = ref(false);
```

### refresh functionality
```js
function refreshPrompt() {
  if (isLoadingFinalPrompt.value) return;

  // visual feedback
  refreshing.value = true;

  // force prompt regeneration
  updateFinalPrompt();

  // reset refreshing state after a short delay
  setTimeout(() => {
    refreshing.value = false;
  }, 2000);
}
```

### button implementation
```vue
<!-- refresh button -->
<button
  @click="refreshPrompt"
  :disabled="isLoadingFinalPrompt"
  :class="[
    'p-2 px-3 rounded-md text-sm flex items-center',
    refreshing ? 'bg-green-600 dark:bg-green-700 text-white' : 'bg-gray-200 text-gray-800 dark:bg-gray-700 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600'
  ]"
  title="regenerate prompt"
>
  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" :class="{'animate-spin': refreshing}">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
  </svg>
  <span class="font-bold">{{ refreshing ? 'refreshing...' : 'refresh' }}</span>
</button>
```

## testing results

all test cases passed successfully:

1. ✅ refresh button appears correctly positioned before the mode buttons
2. ✅ refresh button has the appropriate icon
3. ✅ button shows visual feedback (spinning animation) when clicked
4. ✅ prompt is successfully regenerated when button is clicked
5. ✅ button is properly disabled during prompt loading
6. ✅ styling is consistent with other buttons in both light and dark modes

## screenshots

[not available in this archive]

## related documents

- [task details](../../tasks.md)
- [implementation reflection](../reflection/reflection-refresh-button.md)

## conclusion

this enhancement improves the user experience by providing a quick way to regenerate the prompt without having to manually change settings. the implementation is clean and follows the existing patterns in the codebase.
