package main

type MultiVisitor struct {
	visitors []ResultVisitor
}

func NewMultiVisitor(visitors ...ResultVisitor) *MultiVisitor {
	return &MultiVisitor{visitors}
}

func (visitor *MultiVisitor) Push(inner ResultVisitor) {
	visitor.visitors = append(visitor.visitors, inner)
}

func (visitor *MultiVisitor) Visit(r *Result) {
	for _, v := range visitor.visitors {
		v.Visit(r)
	}
}
