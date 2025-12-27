package offline

import (
	"text/template"
	"context"
	"io"
)

type Pipeline struct {
	PromptTemplate func() (*template.Template, error) // for development we will fetch the prompt template every time
	StoreDocument func(ctx context.Context, place string, document io.Reader) error
}

func (p *Pipeline) Run(
	ctx context.Context,
	place string,
) (err error) {
	// for the given place, we need to
	// 1. collect raw sources (website text, review snippets, menu descriptions,
	//    photo captions)
	// 2. pre-clean (remove html, nav boilerplate, etc)
	// 3. llm-assisted field extraction
	// 4. confidence metadata (how confident are we in the field extraction--do
	//    sources contradict, etc)
	// 5. validation and clamping
	// 6. store in a document database
}