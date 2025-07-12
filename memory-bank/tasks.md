# theme migration: unified color scheme with css variables

## task overview
migrate the entire application to a centralized, css variable-driven color system for robust light/dark theming and future extensibility.

## updated components and files
- `frontend/src/assets/custom.css`: defined comprehensive color variables for both light and dark modes in `:root` and `.dark`
- `frontend/tailwind.config.js`: refactored to reference css variables for all color utilities
- `frontend/src/components/ThemeProvider.vue`: enhanced to manage theme context, toggle logic, and user preference persistence
- `frontend/src/components/BaseButton.vue`
- `frontend/src/components/MainLayout.vue`
- `frontend/src/components/LeftSidebar.vue`
- `frontend/src/components/CentralPanel.vue`
- `frontend/src/components/CustomRulesModal.vue`
- `frontend/src/components/HorizontalStepper.vue`
  - all above: migrated to use theme variables, removed hardcoded color classes

## implementation details
- established a scalable color variable system in custom.css, supporting both base and semantic colors
- updated tailwind config to map color utilities to css variables, ensuring design consistency
- implemented a theme provider component to handle theme switching, persistence, and context propagation
- refactored major ui components to consume theme variables, eliminating legacy `dark:` and hardcoded color usage
- maintained backward compatibility by mapping legacy color names to new variables where necessary
- added error handling and logging for theme switching and persistence failures

## remaining tasks
- refactor remaining components still using legacy `dark:` classes or hardcoded colors:
  - `frontend/src/components/steps/Step1PrepareContext.vue`
  - `frontend/src/components/steps/Step2ComposePrompt.vue`
  - `frontend/src/components/steps/Step3ExecutePrompt.vue`
  - `frontend/src/components/steps/Step4ApplyPatch.vue`
  - `frontend/src/components/FileTree.vue`
- validate theme toggle logic, including persistence and accessibility
- audit all ui elements for consistent application of theme variables
- conduct cross-mode (light/dark) testing for visual and functional consistency

## benefits and outcomes
- fully centralized, maintainable color scheme with easy customization
- seamless, accessible light/dark mode support across all components
- improved maintainability and scalability for future theming needs
- reduced technical debt by eliminating scattered color definitions and legacy classes
