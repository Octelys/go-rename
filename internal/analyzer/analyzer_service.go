package analyzer

import (
	"context"
	"fmt"
	"organizer/internal/abstractions"
	"organizer/internal/ai"
	"sync"
)

type AnalyzerService struct {
	aiProxy    *ai.AiProxy
	pageSource abstractions.PageSource
	context    context.Context
	waitGroup  *sync.WaitGroup
}

func New(
	aiProxy *ai.AiProxy,
	pageSource abstractions.PageSource,
	context context.Context,
	waitGroup *sync.WaitGroup) *AnalyzerService {

	service := AnalyzerService{
		aiProxy:    aiProxy,
		pageSource: pageSource,
		context:    context,
		waitGroup:  waitGroup,
	}

	return &service
}

func (a *AnalyzerService) Analyze() {

	a.waitGroup.Add(1)

	go func() {

		fmt.Println("Analyzer service started.")

		defer a.waitGroup.Done()

		err := a.analyzePages()

		if err != nil {

		}
	}()
}

func (a *AnalyzerService) analyzePages() error {

	for pages := range a.pageSource.Pages() {
		fmt.Printf("Received %d pages to analyze\n", len(pages))
	}

	fmt.Println("Analyzer service stopped.")

	return nil
}
