package entities

type Magazine struct {
	Metadata MagazineMetadata
	Pages    []MagazinePage
	Folder   string
}
