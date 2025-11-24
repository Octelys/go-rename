package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {

	openAiApiKey := os.Getenv("OPENAI_API_KEY")

	if openAiApiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		os.Exit(10000)
	}

	workingDir := os.Getenv("WORKING_DIR")

	if workingDir == "" {
		fmt.Println("WORKING_DIR environment variable is not set")
		os.Exit(10001)
	}

	/* 1. Reads all the file names in the directory */

	files, err := os.ReadDir(workingDir)

	if err != nil {
		fmt.Printf("Unable to read all the files from the working directory: %s\n", err)
		os.Exit(10002)
	}

	fmt.Printf("The following files have been found in the working directory:\n")

	for _, file := range files {
		fmt.Printf("\tFile name: %s\n", file.Name())
	}

	/* 2. Asks the LLM to get the file format used extract the page number from the file name */

	fmt.Printf("\n")
	fmt.Printf("Requesting the assistant to guess the orders of the files... ")

	//	Builds the prompt for the assistant
	var assistantPrompt = strings.Builder{}
	assistantPrompt.WriteString("Below are the files found in the directory. Based on the information found there, sort them according their page number in a JSON array:\n")

	for _, file := range files {
		assistantPrompt.WriteString(file.Name() + "\n")
	}

	//	Initializes the assistant
	client := openai.NewClient(
		option.WithAPIKey(openAiApiKey),
	)

	//	Call the assistant
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(assistantPrompt.String()),
		},
		Model: openai.ChatModelGPT5Mini,
	})

	if err != nil {
		fmt.Printf("[ FAILED ]\n")
		os.Exit(10003)
	}

	var orderedFiles []string
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &orderedFiles)

	if err != nil {
		fmt.Printf("[ FAILED ]\n")
		os.Exit(10004)
	}

	fmt.Printf("[ OK ]\n")

	/* 3. Extract the publication month & year of the first page, assuming it the cover */

}
