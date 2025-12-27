package dsmspaces

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

type IntentsParser struct {
	openAI   openai.Client
	recorder IntentsParseRecorder
}

func NewIntentsParser(
	apiKey string,
	options ...func(*IntentsParser),
) (parser IntentsParser) {
	parser.openAI = openai.NewClient(option.WithAPIKey(apiKey))
	parser.recorder = NullIntentsParseRecorder{}
	for _, option := range options {
		option(&parser)
	}
	return
}

func WithRecorder(recorder IntentsParseRecorder) func(*IntentsParser) {
	return func(p *IntentsParser) {
		p.recorder = recorder
	}
}

func (p *IntentsParser) ParseIntents(
	ctx context.Context,
	query string,
) (intents Intents, err error) {
	var data json.RawMessage
	if data, err = p.ParseIntentsJSON(ctx, query); err != nil {
		return
	}

	if err = json.Unmarshal(data, &intents); err != nil {
		err = fmt.Errorf("parsing intent: unmarshaling completion: %w", err)
		return
	}

	return
}

func (p *IntentsParser) ParseIntentsJSON(
	ctx context.Context,
	query string,
) (intents json.RawMessage, err error) {
	var (
		prompt      = buildSystemPrompt(query)
		temperature = param.NewOpt[float64](0)
		completion  *openai.ChatCompletion
	)
	if completion, err = p.openAI.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: openai.ChatModelGPT3_5Turbo1106,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(prompt),
			},
			Temperature: temperature,
		},
	); err != nil {
		err = fmt.Errorf("parsing intent: %w", err)
		return
	}

	response := completion.Choices[0].Message.Content

	intents = []byte(sanitizeResponse(response))
	if err = ValidateIntents(intents); err != nil {
		err = fmt.Errorf("parsing intent: validating model response: %w", err)
		return
	}
	if err = p.recorder.RecordIntentsParse(
		ctx,
		query,
		intents,
		prompt,
		openai.ChatModelGPT3_5Turbo1106,
		temperature,
	); err != nil {
		err = fmt.Errorf("parsing intent: %w", err)
		return
	}
	return
}

func sanitizeResponse(rsp string) string {
	if strings.HasPrefix(rsp, "```json\n") {
		rsp = strings.TrimPrefix(rsp, "```json\n")
	} else if strings.HasPrefix(rsp, "```\n") {
		rsp = strings.TrimPrefix(rsp, "```\n")
	}

	rsp = strings.TrimSuffix(rsp, "\n```")
	return rsp
}

func buildSystemPrompt(query string) string {
	return strings.Replace(
		strings.Replace(
			systemPrompt,
			"{{INTENTS_SCHEMA}}",
			"```json\n"+intentsSchemaDocument+"\n```",
			1,
		),
		"{{USER_QUERY}}",
		"```\n"+query+"\n```",
		1,
	)
}

const systemPrompt = `
You are an intent parser.

Your job is to convert a user’s natural-language request into a JSON object
that strictly conforms to the provided JSON Schema.

Rules:
- The JSON Schema is authoritative and MUST be followed exactly.
- Only fields defined in the schema may appear in the output.
- If the user does not express an intent corresponding to a schema field,
  that field MUST be omitted.
- Do NOT invent new fields or values.
- Use numeric values in the range [-1.0, 1.0] to represent intent strength:
    1.0  = strong positive desire
    0.0  = neutral / unspecified (omit instead of emitting 0.0)
   -1.0  = strong negative desire / avoidance
- Negative values should be used when the user explicitly avoids something
  (e.g. “no coffee”, “don’t want screens”).
- If the user expresses uncertainty or ambiguity, prefer omitting the field.
- Do NOT include explanatory text, comments, or markdown.
- Output MUST be valid JSON and MUST validate against the schema.

You are NOT allowed to ask follow-up questions.

JSON Schema:
{{INTENTS_SCHEMA}}

User query:
{{USER_QUERY}}

Produce a single JSON object that conforms to the schema and represents
the user’s intent. Return the JSON as plain text, starting with "{" and ending
with "}". The result will be parsed strictly as JSON, so exclude any markdown
code fences or similar.
`
