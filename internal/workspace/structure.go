package workspace

type File struct {
	Content string `json:"content"`
}

type Dir struct {
	Entries map[string]*Entry `json:"entries"`
}

type Entry struct {
	File *File `json:"file"`
	Dir  *Dir  `json:"dir"`
}
