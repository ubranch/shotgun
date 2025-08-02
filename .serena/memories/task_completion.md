# Task Completion Guidelines

## When a task is completed:

1. **Test the changes** - Run `wails dev` to test in development mode
2. **Check for errors** - Monitor console for JavaScript errors and Go runtime errors
3. **Verify functionality** - Test the specific feature that was modified
4. **Performance check** - Ensure no performance regressions, especially for:
   - Context generation
   - File tree operations
   - Prompt updates
   - Token counting

## Common issues to watch for:
- Debouncing not working properly causing excessive API calls
- Memory leaks from event listeners not being cleaned up
- Reactive watchers triggering too frequently
- Large file operations blocking the UI
- Context generation taking too long

## Testing workflow:
1. Start with `wails dev`
2. Test the main user flow: select folder → generate context → compose prompt → execute
3. Test edge cases like large projects, empty projects, error conditions
4. Verify responsive design on different screen sizes