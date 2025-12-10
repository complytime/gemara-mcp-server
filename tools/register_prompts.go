package tools

import (
	"context"

	"github.com/complytime/gemara-mcp-server/tools/prompts"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerPrompts registers all prompts with the server
func (g *GemaraAuthoringTools) registerPrompts() []server.ServerPrompt {
	var prompts []server.ServerPrompt

	prompts = append(prompts, g.newCreateLayer1GuidancePrompt())
	prompts = append(prompts, g.newCreateLayer2ControlsPrompt())
	prompts = append(prompts, g.newCreateLayer3PoliciesPrompt())
	prompts = append(prompts, g.newGemaraQuickStartPrompt())

	return prompts
}

func (g *GemaraAuthoringTools) newCreateLayer1GuidancePrompt() server.ServerPrompt {
	return server.ServerPrompt{
		Prompt: mcp.NewPrompt(
			"create-layer1-guidance",
			mcp.WithPromptDescription("Guide for creating Layer 1 Guidance documents. Provides YAML structure, examples, and best practices."),
		),
		Handler: func(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return mcp.NewGetPromptResult(
				"Creating Layer 1 Guidance Documents",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompts.CreateLayer1Prompt)),
				},
			), nil
		},
	}
}

func (g *GemaraAuthoringTools) newCreateLayer2ControlsPrompt() server.ServerPrompt {
	return server.ServerPrompt{
		Prompt: mcp.NewPrompt(
			"create-layer2-controls",
			mcp.WithPromptDescription("Guide for creating Layer 2 Control Catalogs. Provides YAML structure, examples, and best practices."),
		),
		Handler: func(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return mcp.NewGetPromptResult(
				"Creating Layer 2 Control Catalogs",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompts.CreateLayer2Prompt)),
				},
			), nil
		},
	}
}

func (g *GemaraAuthoringTools) newCreateLayer3PoliciesPrompt() server.ServerPrompt {
	return server.ServerPrompt{
		Prompt: mcp.NewPrompt(
			"create-layer3-policies",
			mcp.WithPromptDescription("Guide for creating Layer 3 Policy documents. Provides YAML structure, examples, and best practices."),
		),
		Handler: func(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return mcp.NewGetPromptResult(
				"Creating Layer 3 Policy Documents",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompts.CreateLayer3Prompt)),
				},
			), nil
		},
	}
}

func (g *GemaraAuthoringTools) newGemaraQuickStartPrompt() server.ServerPrompt {
	return server.ServerPrompt{
		Prompt: mcp.NewPrompt(
			"gemara-quick-start",
			mcp.WithPromptDescription("Quick start guide for creating your first Gemara artifacts. Provides step-by-step instructions and common workflows."),
		),
		Handler: func(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return mcp.NewGetPromptResult(
				"Gemara Quick Start Guide",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompts.QuickStartPrompt)),
				},
			), nil
		},
	}
}
