package dedup

type Group struct {
	Size               int64
	FirstBlockChecksum uint32
	FinalBlockChecksum uint32
	Paths              []string
}
