package dsmspaces

import (
	"context"
	"dsmspaces/pkg/logger"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type IntentParser struct {
	openAI openai.Client
	cache  map[string]map[Attr]float64 // TODO: use a shared cache
}

func NewIntentParser(apiKey string) (parser IntentParser) {
	parser.openAI = openai.NewClient(option.WithAPIKey(apiKey))
	parser.cache = map[string]map[Attr]float64{}
	return
}

func (p *IntentParser) ParseIntent(
	ctx context.Context,
	query string,
) (expectations map[Attr]float64, err error) {
	var exists bool
	if expectations, exists = p.cache[query]; exists {
		return
	}

	var (
		prompt     = fmt.Sprintf(promptTemplate, query)
		completion *openai.ChatCompletion
	)
	if completion, err = p.openAI.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: openai.ChatModelChatgpt4oLatest,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(prompt),
			},
		},
	); err != nil {
		err = fmt.Errorf("parsing intent: %w", err)
		return
	}

	logger.Get(ctx).Debug(
		"parsing intent",
		"query", query,
		"completion", completion.Choices[0].Message.Content,
	)

	if err = json.Unmarshal(
		[]byte(completion.Choices[0].Message.Content),
		&expectations,
	); err != nil {
		err = fmt.Errorf("parsing intent: unmarshaling completion: %w", err)
		return
	}

	p.cache[query] = expectations
	return
}

const promptTemplate = `Convert the user query into an expectation vector.
Use values from -1.0 to 1.0:
- 1.0 = must have
- 0.5 = preference
- 0.0 = indifferent
- -0.5 = prefer absence
- -1.0 = must not have

Attributes: affordable, cozy, quiet, coffee, alcohol, eveningFriendly, readingFriendly, classy, screenless, spacious
Example 1:
Query: "I want a cozy place to read, no coffee"
Output:
{
  "cozy": 1.0,
  "quiet": 0.8,
  "evening": 0.6,
  "coffee": -1.0,
  "readingFriendly": 0.9
}

Example 2:
Query: "Quiet bar for reading at night"
Output:
{
  "cozy": 0.7,
  "eveningFriendly": 1.0,
  "readingFriendly": 0,
  "alcohol": 1.0,
}

Query: %s
Output:`
