package ai

// Available models on OpenRouter
var AvailableModels = map[string]string{
    "deepseek-coder": "deepseek/deepseek-coder:33b-instruct",   
    "deepseek-chat":  "deepseek/deepseek-chat",                 
    "claude-3-sonnet": "anthropic/claude-3-sonnet:beta",         
    "gpt-3.5":        "openai/gpt-3.5-turbo",                   
    "gpt-4":          "openai/gpt-4",                           
}

type ModelConfig struct {
    Name         string
    Provider     string
    ContextSize  int
    SupportsJSON bool
}

var ModelConfigs = map[string]ModelConfig{
    "deepseek/deepseek-coder:33b-instruct": {
        Name:         "DeepSeek Coder",
        Provider:     "DeepSeek",
        ContextSize:  16384,
        SupportsJSON: true,
    },
    "deepseek/deepseek-chat": {
        Name:         "DeepSeek Chat",
        Provider:     "DeepSeek", 
        ContextSize:  32768,
        SupportsJSON: true,
    },
}