package workspace

type file struct {
	Content string
}

type dir struct {
	Entries map[string]*entry
}

type entry struct {
	File *file
	Dir  *dir
}
