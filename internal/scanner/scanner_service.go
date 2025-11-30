package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"organizer/internal/abstractions/entities"
	"organizer/internal/ai"
	"organizer/internal/configuration"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	AssistantPrompt = "Below are the files found in the directory. Based on the information found there, sort them according to their scanner number in a JSON array (for example: [{\"file\": \"page_01.pdf\", \"number\": 1 }, {\"file\": \"page_02.pdf\", \"number\": 2 }]). If the 1st file starts at the number 0, make sure you start counting at 1. Return only valid JSON and no extra text. Exclude the duplicate file names, especially those who have an extra ' 1', and exclude the files that does not seem to be entirely different from the rest. Make sure the first page is number 1."
)

type ScannerService struct {
	workingDirectory     string
	aiProxy              *ai.AiProxy
	context              context.Context
	magazinePagesChannel chan entities.MagazinePages
	waitGroup            *sync.WaitGroup
}

func New(
	configurationService *configuration.ConfigurationService,
	aiProxy *ai.AiProxy,
	context context.Context,
	waitGroup *sync.WaitGroup) *ScannerService {

	service := ScannerService{
		workingDirectory:     configurationService.WorkingDirectory,
		context:              context,
		aiProxy:              aiProxy,
		waitGroup:            waitGroup,
		magazinePagesChannel: make(chan entities.MagazinePages),
	}

	return &service
}

func (s *ScannerService) Scan() {

	s.waitGroup.Add(1)

	go func() {

		fmt.Println("Scanner service started.")

		defer s.waitGroup.Done()

		err := s.readFolders()
		if err != nil {

		}
	}()
}

func (s *ScannerService) readFolders() error {

	folders, err := os.ReadDir(s.workingDirectory)

	if err != nil {
		return fmt.Errorf("unable to read all the folder from the working directory: %s", err)
	}

	for _, folder := range folders {

		if !folder.IsDir() {
			continue
		}

		publicationFolder := filepath.Join(s.workingDirectory, folder.Name())

		files, err := os.ReadDir(publicationFolder)

		if err != nil {
			return fmt.Errorf("unable to read all the files from the directory: %s", err)
		}

		fmt.Printf("Analyzing folder '%s'\n", folder.Name())

		var assistantPrompt strings.Builder
		assistantPrompt.WriteString(AssistantPrompt)
		assistantPrompt.WriteString("\n")

		for _, file := range files {
			assistantPrompt.WriteString(file.Name())
			assistantPrompt.WriteString("\n")
		}

		aiResponse, err := s.aiProxy.SendRequest(assistantPrompt.String())

		if err != nil {
			return err
		}

		var orderedPages []entities.MagazinePage
		if err := json.Unmarshal([]byte(aiResponse), &orderedPages); err != nil {
			return fmt.Errorf("Unable to retrieve the ordered pages from the assistant: %v\n", err)
		}

		fmt.Printf("Found %d pages in folder '%s'\n", len(orderedPages), folder.Name())

		magazinePages := entities.MagazinePages{
			Pages:  orderedPages,
			Folder: publicationFolder,
		}

		s.magazinePagesChannel <- magazinePages
	}

	close(s.magazinePagesChannel)

	fmt.Println("Scanner service stopped.")

	return nil
}

func (s *ScannerService) Pages() <-chan entities.MagazinePages {
	return s.magazinePagesChannel
}
