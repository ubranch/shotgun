# memory bank: tasks

## current task

task: add line count and size information to generated project context

## task details

add line count and file size information (in kb) to the first step under the "generated project context:" title. this should display the total number of lines and the total size in kilobytes of the generated context.

## task status

- [x] analyze task requirements
- [x] determine task complexity
- [x] create implementation plan
- [x] execute implementation
- [x] test functionality
- [x] document changes

## complexity level

level 2: simple enhancement

## analysis outcome

after examining the codebase, i've identified that this task requires modifying the `step1preparecontext.vue` component to display statistics about the generated context.

key findings:
1. the component already receives the generated context via the `generatedcontext` prop
2. the component displays this context in a textarea with a header section containing the title "generated project context:"
3. the ui includes a section where we can add the statistics information (line count and size in kb)
4. the implementation will require calculating:
   - total number of lines in the context
   - total size of the context in kilobytes
5. this information needs to be displayed in the first step interface, specifically under the "generated project context:" title

this is a level 2 enhancement as it involves:
- modifying a single component with straightforward implementation
- adding simple calculations for line count and file size
- displaying the information in the existing ui
- no backend changes required (context data is already available in the frontend)

## implementation plan

### step 1: add computed properties for context statistics

add computed properties to the `step1preparecontext.vue` component that calculate the line count and file size:

```javascript
// computed properties for context statistics
const contextStats = computed(() => {
  if (!props.generatedContext) return { lines: 0, sizeKb: 0 };

  const lines = props.generatedContext.split('\n').length;
  const sizeKb = (props.generatedContext.length / 1024).toFixed(1);

  return { lines, sizeKb };
});
```

### step 2: modify the header section to display statistics

update the header section to include the line count and file size information:

```html
<div class="flex justify-between items-center mb-2">
  <div>
    <h3 class="text-md font-medium text-gray-700 dark:text-gray-300">
      generated project context:
    </h3>
    <p v-if="generatedContext" class="text-xs text-gray-500 dark:text-gray-400">
      {{ contextStats.lines }} lines ({{ contextStats.sizeKb }} kb)
    </p>
  </div>
  <button
    v-if="generatedContext"
    @click="copyGeneratedContextToClipboard"
    class="px-3 py-2 bg-light-accent dark:bg-dark-accent text-white text-sm font-semibold rounded-md hover:bg-light-accent-hover dark:hover:bg-dark-accent-hover focus:outline-none disabled:bg-gray-300 dark:disabled:bg-gray-700 flex items-center gap-1"
    :class="{
      'bg-green-600 dark:bg-green-700': copySuccess,
    }"
  >
    <!-- button content remains the same -->
    <svg
      v-if="!copySuccess"
      xmlns="http://www.w3.org/2000/svg"
      class="h-4 w-4"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
      />
    </svg>
    <svg
      v-else
      xmlns="http://www.w3.org/2000/svg"
      class="h-4 w-4"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M5 13l4 4L19 7"
      />
    </svg>
    {{ copyButtonText }}
  </button>
</div>
```

### step 3: handle edge cases and testing

1. ensure the statistics are only displayed when context is actually available
2. handle large contexts efficiently to avoid performance issues
3. update the display when context changes (this is automatically handled by the computed property)
4. add checks to avoid errors when the context is null or undefined
5. test with various project sizes to ensure accurate display

### step 4: testing plan

test the implementation with:
1. small projects (few files)
2. medium projects (dozens of files)
3. large projects (hundreds of files)
4. verify line count accuracy matches actual line breaks in the text
5. verify file size calculation is accurate

### step 5: validation

verify that the statistics:
1. appear in the correct location (under the "generated project context:" title)
2. update correctly when a new project is loaded or context is refreshed
3. display accurate information
4. maintain proper formatting in both light and dark themes

## task completion

task completed successfully. the implementation adds line count and file size information to the first step under the "generated project context:" title, enhancing the user experience with valuable metadata.

key accomplishments:
1. added the `contextStats` computed property to calculate lines and size in kb
2. modified the ui to display the statistics below the heading
3. tested with different project sizes to verify the accuracy
4. verified that the statistics update correctly when the context changes
5. confirmed that the display works properly in both light and dark modes
6. created complete documentation:
   - created reflection document in memory-bank/reflection/reflection-context-statistics.md
   - created archive document in memory-bank/archive/archive-context-statistics.md

the enhancement now displays the number of lines and size in kilobytes directly under the "generated project context:" title, providing users with useful information about the context they're working with.

## completed tasks

### task: add refresh button before the prompt modes

#### task details

add a "refresh" button with a refresh icon before the "prompt" button in the modes section. when pressed, the button should invoke the regeneration of the prompt. the button should be positioned to the left of the existing mode buttons.

#### task status

- [x] review current ui components and identify where to add the refresh button
- [x] analyze project complexity for the new enhancement
- [x] create detailed plan for implementing the refresh button
- [x] identify files that need modification
- [x] implement the refresh button functionality
- [x] test the functionality
- [x] document the changes

#### complexity level

level 2: simple enhancement

#### analysis outcome

the shotgun application is a desktop tool built with wails (go backend, vue frontend) that helps users with prompt engineering and testing. the application follows a structured 4-step workflow:

1. **prepare context**: select project folder and generate context
2. **compose prompt**: craft prompts using the project context
3. **execute prompt**: process the ai-generated code changes
4. **apply patch**: implement the generated changes

for this enhancement, we need to modify the step2composeprompt.vue component which contains the mode buttons (prompt, plan, build, bug, reflect). the mode buttons are implemented as a v-for loop that iterates over prompttemplates, displaying each with its icon and short name.

key findings:

1. the mode buttons are in step2composeprompt.vue in a button group
2. each button has an icon and text label
3. button selection updates selectedprompttemplekey and triggers prompt generation
4. the refresh functionality should re-trigger prompt generation using updatefinalprompt()
5. button styling should match existing buttons for consistency

this is a level 2 enhancement as it involves modifying a single component with a straightforward implementation.

#### implementation plan

see archive-refresh-button.md for full implementation details.

#### task completion

this task has been successfully completed. the refresh button functionality has been implemented and tested. the enhancement improves the user experience by providing a quick way to regenerate the prompt without having to manually change settings.
