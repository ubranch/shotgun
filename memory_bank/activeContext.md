# active context

## current mode

- **mode**: van
- **status**: in progress
- **started**: initial setup
- **next mode**: implement (level 1 task)

## current task

analyzing the shotgun project structure and understanding the Step2ComposePrompt.vue component which has uncommitted changes according to git status.

## findings

- project is a prompt engineering tool named "shotgun"
- built with go backend (wails framework) and vue.js frontend
- Step2ComposePrompt.vue component is responsible for composing prompts with template support
- component features:
  - template selection (dev, architect, findBug, projectManager, promptEnhancer)
  - token counting using google gemini api
  - custom rules modal
  - clipboard integration
  - user task input
  - file list context integration

## complexity assessment

- **level**: 1 (quick bug fix)
- **scope**: focused on single component
- **risk**: low (understanding phase)
- **next steps**: narrate plan and transition to implement mode
