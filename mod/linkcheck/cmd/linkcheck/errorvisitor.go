package main

import "net/http"

type ErrorVisitor struct {
	inner ResultVisitor
}

func NewErrorsOnlyVisitor() *ErrorVisitor {
	return &ErrorVisitor{}
}

func (visitor *ErrorVisitor) SetInner(inner ResultVisitor) *ErrorVisitor {
	visitor.inner = inner
	return visitor
}

func (visitor *ErrorVisitor) Visit(r *Result) {
	if visitor.inner != nil &&
		(r.URLParseError != nil ||
			r.NetworkError != nil ||
			r.StatusCode != http.StatusOK) {
		visitor.inner.Visit(r)
	}
}
