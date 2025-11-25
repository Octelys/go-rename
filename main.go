package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	//"github.com/openai/openai-go/v3/responses"
)

func main() {
	ctx := context.Background()

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

	// 1. Read all the file names in the directory
	fmt.Printf("Reading the name of the files from the working directory... ")

	files, err := os.ReadDir(workingDir)
	if err != nil {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tUnable to read all the files from the working directory: %s\n", err)
		os.Exit(10002)
	}

	fmt.Printf("\rFound %d files from the working directory [ OK ]\t\t\n", len(files))

	// 2. Ask the LLM to infer file order from file names
	fmt.Printf("Requesting the assistant to guess the order of the files... ")

	var assistantPrompt strings.Builder
	assistantPrompt.WriteString("Below are the files found in the directory. Based on the information found there, sort them according to their page number in a JSON array (for example: [\"page_01.pdf\", \"page_02.pdf\"]). Return only valid JSON and no extra text.\n")

	for _, file := range files {
		assistantPrompt.WriteString(file.Name())
		assistantPrompt.WriteString("\n")
	}

	client := openai.NewClient(
		option.WithAPIKey(openAiApiKey),
	)

	chatResp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT5Mini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(assistantPrompt.String()),
		},
	})

	if err != nil {
		fmt.Printf("[ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tError while ordering files: %v\n", err)
		os.Exit(10003)
	}

	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		fmt.Printf("[ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tAssistant returned an empty response when ordering files\n")
		os.Exit(10004)
	}

	var orderedFiles []string
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &orderedFiles); err != nil {
		fmt.Printf("[ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tUnable to parse assistant JSON response: %v\n", err)
		os.Exit(10005)
	}

	if len(orderedFiles) == 0 {
		fmt.Printf("[ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tThe assistant did not return any ordered files\n")
		os.Exit(10006)
	}

	fmt.Printf("[ OK ]\n")

	// 3. Extract the publication month & year of the first page, assuming it is the cover
	coverFileName := orderedFiles[0]

	fmt.Printf("Analyzing cover file '%s'... ", coverFileName)

	coverPath := filepath.Join(workingDir, coverFileName)

	if _, err := os.Stat(coverPath); err != nil {
		fmt.Printf("[ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tCover file '%s' does not exist or is not accessible: %v\n", coverPath, err)
		os.Exit(10007)
	}

	reader, err := os.Open(coverPath)

	if err != nil {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tCover file '%s' does not exist or is not accessible: %v\n", coverPath, err)
		os.Exit(10007)
	}

	//defer reader.Close()

	file, err := client.Files.New(context.TODO(), openai.FileNewParams{
		File:    reader,
		Purpose: openai.FilePurposeUserData,
	})

	if err != nil {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tUnable to create a new file: %v\n", err)
		os.Exit(10008)
	}

	storeName := "go-runner-vector-store"

	vectorStore, err := client.VectorStores.New(
		ctx,
		openai.VectorStoreNewParams{
			ExpiresAfter: openai.VectorStoreNewParamsExpiresAfter{
				Days: 1,
			},
			Name: openai.String(storeName),
		},
	)

	if err != nil {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tUnable to create a vector store: %v\n", err)
		os.Exit(10008)
	}

	batch, err := client.VectorStores.FileBatches.New(
		ctx,
		vectorStore.ID,
		openai.VectorStoreFileBatchNewParams{
			FileIDs: []string{file.ID},
		},
	)

	if err != nil {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tUnable to upload the file to the vector store: %v\n", err)
		os.Exit(10009)
	}

	//	Watchout: the version I use has a bug where the batch ID and vector store ID are swapped
	polledBatch, err := client.VectorStores.FileBatches.PollStatus(
		ctx,
		batch.ID,
		vectorStore.ID,
		0,
	)

	if err != nil {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tUnable to upload the file to the vector store: %v\n", err)
		os.Exit(10009)
	}

	if polledBatch.Status != "completed" {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("\tUnable to upload the file to the vector store: status is %v\n", polledBatch.Status)
		os.Exit(10009)
	}

	assistantPrompt.Reset()
	assistantPrompt.WriteString("You are given a JPG file containing an image of a cover page of a French publication: ")
	assistantPrompt.WriteString(coverFileName)
	assistantPrompt.WriteString(". Based on typical naming conventions and any context you can infer, ")
	assistantPrompt.WriteString("return only the publication month and year in the format `MMMM YYYY` (for example: `Juin 2024`). ")
	assistantPrompt.WriteString("If you cannot determine it, answer exactly `Unknown`. Do not add any extra explanation.")

	fileSearchToolParam := responses.FileSearchToolParam{
		VectorStoreIDs: []string{vectorStore.ID},
	}

	toolParam := responses.ToolUnionParam{
		OfFileSearch: &fileSearchToolParam,
	}

	publicationDateResponse, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(assistantPrompt.String()),
		},
		Model: openai.ChatModelGPT5,
		Tools: []responses.ToolUnionParam{
			toolParam,
		},
	})

	if err != nil {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("Unable to analyze the cover: %v\n", err)
		os.Exit(10008)
	}

	if publicationDateResponse == nil || publicationDateResponse.OutputText() == "" {
		fmt.Printf(" [ FAILED ]\n")
		fmt.Printf("\n")
		fmt.Printf("Model did not return a publication date\n")
		os.Exit(10009)
	}

	fmt.Printf(" [ OK ]\n")
	fmt.Printf("\n")

	publicationDate := strings.TrimSpace(publicationDateResponse.OutputText())
	fmt.Printf("Publication date (month & year): %s\n", publicationDate)
}
