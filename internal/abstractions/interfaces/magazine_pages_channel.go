package interfaces

import (
	"organizer/internal/abstractions/entities"
)

type MagazinePagesChannel interface {
	Pages() <-chan entities.MagazinePages
}
