package abstractions

type PageSource interface {
	Pages() <-chan []Page
}
