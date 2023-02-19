package main

type CountVisitor struct {
	count int
}

func (visitor *CountVisitor) GetCount() int { return visitor.count }

func (visitor *CountVisitor) Visit(r *Result) {
	visitor.count++
}
