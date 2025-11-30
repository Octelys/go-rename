package abstractions

type Publication struct {
	Title  string  `json:"title"`
	Number uint8   `json:"number"`
	Month  []uint8 `json:"months"`
	Year   uint16  `json:"year"`
}
