package workspace

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Dir struct {
	Name    string            `json:"name"`
	Entries map[string]*Entry `json:"entries"`
}

type Entry struct {
	File *File `json:"file"`
	Dir  *Dir  `json:"dir"`
}

// TODO: enforce oneof?
