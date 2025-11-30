package configuration

import (
	"fmt"
	"os"
)

const (
	OpenaiApiKeyEnvVarName     = "OPENAI_API_KEY"
	WorkingDirectoryEnvVarName = "WORKING_DIR"
)

type ConfigurationService struct {
	OpenAiApiKey     string
	WorkingDirectory string
}

func New() (*ConfigurationService, error) {

	openAiApiKey := os.Getenv(OpenaiApiKeyEnvVarName)
	if openAiApiKey == "" {
		return nil, fmt.Errorf("%s environment variable is not set", OpenaiApiKeyEnvVarName)
	}

	workingDir := os.Getenv(WorkingDirectoryEnvVarName)
	if workingDir == "" {
		return nil, fmt.Errorf("%s environment variable is not set", OpenaiApiKeyEnvVarName)
	}

	configurationService := ConfigurationService{
		OpenAiApiKey:     openAiApiKey,
		WorkingDirectory: workingDir,
	}

	return &configurationService, nil
}
