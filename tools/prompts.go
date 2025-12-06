package tools

import (
	_ "embed"
)

//go:embed prompts/create-layer1.md
var createLayer1Prompt string

//go:embed prompts/create-layer2.md
var createLayer2Prompt string

//go:embed prompts/create-layer3.md
var createLayer3Prompt string

//go:embed prompts/quick-start.md
var quickStartPrompt string
