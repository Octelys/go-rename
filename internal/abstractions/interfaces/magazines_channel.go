package interfaces

import (
	"organizer/internal/abstractions/entities"
)

type MagazinesChannel interface {
	Magazines() <-chan entities.Magazine
}
