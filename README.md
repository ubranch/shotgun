# shotgun

shotgun is a tool for streamlined prompt engineering and testing. it provides a user-friendly interface for crafting prompts, managing context, and executing prompts with ai models.

## requirements

- go 1.24.0 or higher
- node.js and pnpm for frontend development
- google api key for gemini token counting

## setup

### google api key for gemini token counting

to use the accurate gemini token counting feature, you need to set up your google ai api key:

1. obtain an api key from [google ai studio](https://ai.google.dev/)
2. set the environment variable before running the application:

on windows (powershell):
```
$env:GOOGLE_API_KEY = "your-api-key"
```

on unix-like systems (bash):
```
export GOOGLE_API_KEY="your-api-key"
```

note: the application will fall back to displaying an error message if the api key is not set or invalid.

## development

### running in development mode

```
wails dev
```

### building

```
wails build
```

## features

- accurate gemini token counting using the official google ai sdk
- file tree visualization with gitignore support
- customizable prompt templates
- live token counting and validation
- clipboard integration
