package fileset

type FileSet map[string]struct{}

func New() FileSet {
	return FileSet{}
}

var nod = struct{}{}

func (I *FileSet) Add(filename string) {
	(*I)[filename] = nod
}

func (I *FileSet) Remove(filename string) {
	delete(*I, filename)
}

func (I FileSet) IsEmpty() bool {
	return len(I) == 0
}
