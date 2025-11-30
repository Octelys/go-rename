package ai

import (
	"context"
	"fmt"

	"organizer/internal/configuration"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type AiProxy struct {
	client  *openai.Client
	context context.Context
}

func New(
	configurationService *configuration.ConfigurationService,
	context context.Context) (*AiProxy, error) {

	openaiClient := openai.NewClient(
		option.WithAPIKey(configurationService.OpenAiApiKey),
	)

	return &AiProxy{
		client:  &openaiClient,
		context: context,
	}, nil
}

func (aiProxy *AiProxy) SendRequest(assistantPrompt string) (string, error) {

	response, err := aiProxy.client.Responses.New(aiProxy.context, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: []responses.ResponseInputItemUnionParam{
				{
					OfInputMessage: &responses.ResponseInputItemMessageParam{
						Role: "user",
						Content: responses.ResponseInputMessageContentListParam{
							{
								OfInputText: &responses.ResponseInputTextParam{
									Text: assistantPrompt,
								},
							},
						},
					},
				},
			},
		},
		Model: openai.ChatModelGPT5Mini,
	})

	if err != nil {
		return "", fmt.Errorf("unable to process the prompt: %v", err)
	}

	outputText := response.OutputText()

	return outputText, nil
}
