package fs

type InoSet map[Ino]struct{}

func NewInoSet() InoSet {
	return make(InoSet)
}

func (set InoSet) Add(ino Ino) {
	set[ino] = struct{}{}
}

func (set InoSet) Exists(ino Ino) bool {
	_, exists := set[ino]
	return exists
}
