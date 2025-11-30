package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"organizer/internal/abstractions/entities"
	"organizer/internal/abstractions/interfaces"
	"organizer/internal/ai"
	"os"
	"path/filepath"
	"sync"
)

const (
	AssistantPrompt = "You are given a JPG file containing an image of a cover scanner of a French publication. Based on typical naming conventions and any context you can infer, return only the title, publication number and publication month and year in the JSON format `{ \"title\": string, \"months\": [number,], \"year\": number, \"number\": number }`. If you cannot determine it, answer exactly `Unknown`. Do not add any extra explanation."
)

type AnalyzerService struct {
	aiProxy              *ai.AiProxy
	magazinePagesChannel interfaces.MagazinePagesChannel
	magazinesChannel     chan entities.Magazine
	context              context.Context
	waitGroup            *sync.WaitGroup
}

func New(
	aiProxy *ai.AiProxy,
	magazinePagesChannel interfaces.MagazinePagesChannel,
	context context.Context,
	waitGroup *sync.WaitGroup) *AnalyzerService {

	service := AnalyzerService{
		aiProxy:              aiProxy,
		magazinePagesChannel: magazinePagesChannel,
		magazinesChannel:     make(chan entities.Magazine),
		context:              context,
		waitGroup:            waitGroup,
	}

	return &service
}

func (a *AnalyzerService) Run() {

	a.waitGroup.Add(1)

	go func() {

		fmt.Println("Analyzer service started.")

		defer a.waitGroup.Done()

		err := a.monitor()

		if err != nil {

		}
	}()
}

func (a *AnalyzerService) monitor() error {

	for magazinePages := range a.magazinePagesChannel.Pages() {
		fmt.Printf("Received magazine with %d pages to analyze\n", len(magazinePages.Pages))
		a.analyzePages(magazinePages)
	}

	close(a.magazinesChannel)

	fmt.Println("Analyzer service stopped.")

	return nil
}

func (a *AnalyzerService) analyzePages(magazinePages entities.MagazinePages) {

	if magazinePages.Pages == nil || len(magazinePages.Pages) == 0 {
		fmt.Println("No pages to analyze.")
		return
	}

	coverFileName := magazinePages.Pages[0].File

	fmt.Printf("Analyzing cover file '%s'\n", coverFileName)

	coverPath := filepath.Join(magazinePages.Folder, coverFileName)

	if _, err := os.Stat(coverPath); err != nil {
		fmt.Printf("Cover file '%s' does not exist or is not accessible: %v\n", coverPath, err)
		return
	}

	reader, err := os.Open(coverPath)

	if err != nil {
		fmt.Printf("Cover file '%s' does not exist or is not accessible: %v\n", coverPath, err)
		return
	}

	defer reader.Close()

	response, err := a.aiProxy.SendRequestWithImage(AssistantPrompt, reader)

	if err != nil {
		fmt.Printf("An error occurred trying to analyze the cover file '%s': %v\n", coverPath, err)
		return
	}

	if response == "" || response == "Unknown" {
		fmt.Printf("Unable to retieve the metadata of the cover file '%s'\n", coverPath)
		return
	}

	var metadata entities.MagazineMetadata
	if err := json.Unmarshal([]byte(response), &metadata); err != nil {
		fmt.Printf("Unable to decode the magazine metadata of cover file '%s': %v\n", coverPath, err)
		fmt.Printf("Received: %s\n", response)
		return
	}

	fmt.Printf("Publication %s #%d\n", metadata.Title, metadata.Number)

	a.magazinesChannel <- entities.Magazine{
		Metadata: metadata,
		Pages:    magazinePages.Pages,
		Folder:   magazinePages.Folder,
	}
}

func (a *AnalyzerService) Magazines() <-chan entities.Magazine {
	return a.magazinesChannel
}
