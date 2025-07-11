# memory bank: style guide

## code formatting

- **go**: standard go formatting (`gofmt`)
- **vue/js/ts**: 2 space indentation, semicolons
- **css**: tailwind classes, grouped by functionality

## naming conventions

- **go**: camelCase for exported functions/variables, lowercase for package-private
- **vue components**: pascal case (e.g., `FileTree.vue`)
- **js/ts functions**: camel case
- **css classes**: kebab-case

## documentation standards

- go functions should have descriptive comments following godoc format
- vue components should have top-level description comments
- complex logic should include inline comments explaining the approach
- public apis should have clear documentation

## commit message format

preferred format: conventional commits
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

types: feat, fix, docs, style, refactor, perf, test, build, ci, chore
