package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type AgentResponse struct {
	Type string `json:"type" jsonschema:"enum=tool,enum=ask,enum=final"`
	Tool string `json:"tool" jsonschema:"enum=,enum=getContext,enum=setContext,enum=getPods,enum=getNamespaces,enum=describePod,enum=getPodLogs"`

	Input    map[string]string `json:"input,omitempty"`
	Question string            `json:"question"`
	Answer   string            `json:"answer"`
}

type LLM interface {
	Call(messages []Message) (AgentResponse, error)
}

type OpenAIBranin struct {
	cli openai.Client
}

func NewOpenAIBrain() *OpenAIBranin {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		fmt.Print("[ERROR] missing OPENAI_API_KEY")
		os.Exit(1)
	}
	client := openai.NewClient(
		option.WithAPIKey(openAIKey),
	)
	return &OpenAIBranin{
		cli: client,
	}
}

func GenerateSchema[T any]() map[string]any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	var v T
	schema := reflector.Reflect(v)

	data, _ := json.Marshal(schema)

	var result map[string]any
	_ = json.Unmarshal(data, &result)

	return result
}

var agentResponseSchema = GenerateSchema[AgentResponse]()

func (b *OpenAIBranin) Call(messages []Message) (AgentResponse, error) {
	input := buildPrompt(messages)
	fmt.Println("HERE1")
	params := responses.ResponseNewParams{
		Model: openai.ChatModelGPT5_2,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(input),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
					Name:   "agent_response",
					Schema: agentResponseSchema,
					Strict: openai.Bool(true),
				},
			},
		},
	}
	resp, err := b.cli.Responses.New(context.Background(), params)
	if err != nil {
		fmt.Println("ERROR:", err)
		return AgentResponse{}, err
	}
	raw := strings.TrimSpace(resp.OutputText())

	// for debugging
	fmt.Println("RAW:", raw)

	dec := json.NewDecoder(strings.NewReader(raw))

	var agentResp AgentResponse
	if err := dec.Decode(&agentResp); err != nil {
		return AgentResponse{}, err
	}

	return agentResp, nil
}

func buildPrompt(messages []Message) string {
	content, err := os.ReadFile("internal/prompts/system.txt")
	if err != nil {
		panic(err)
	}

	prompt := string(content) + "\n\nConversation:\n"

	for _, m := range messages {
		prompt += m.Role + ": " + m.Content + "\n"
	}

	return prompt
}
