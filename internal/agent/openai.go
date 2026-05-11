package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"simple-k8s-agent/internal/prompts"

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
		Tools: []responses.ToolUnionParam{
			{
				OfFunction: &responses.FunctionToolParam{
					Name:        "getPods",
					Description: openai.String("Get list of the pods in namespace"),
					Strict:      openai.Bool(true),
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"namespace": map[string]string{
								"type": "string",
							},
						},
						"required":             []string{"namespace"},
						"additionalProperties": false,
					},
				},
			},
			{
				OfFunction: &responses.FunctionToolParam{
					Name:        "getNamespaces",
					Description: openai.String("Get list of the namespaces"),
					Strict:      openai.Bool(true),
					Parameters: map[string]any{
						"type":                 "object",
						"properties":           map[string]any{},
						"required":             []string{},
						"additionalProperties": false,
					},
				},
			},
			{
				OfFunction: &responses.FunctionToolParam{
					Name:        "describePod",
					Description: openai.String("Describe pod in namespace"),
					Strict:      openai.Bool(true),
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"pod": map[string]string{
								"type": "string",
							},
							"namespace": map[string]string{
								"type": "string",
							},
						},
						"required":             []string{"namespace", "pod"},
						"additionalProperties": false,
					},
				},
			},
		},
	}
	response, err := b.cli.Responses.New(context.Background(), params)
	if err != nil {
		fmt.Println("ERROR:", err)
		return AgentResponse{}, err
	}
	// raw := strings.TrimSpace(resp.OutputText())
	for _, item := range response.Output {
		if item.Type == "function_call" {
			toolCall := item.AsFunctionCall()
			if toolCall.Name == "get_weather" {
				// Extract arguments and call your function
				var args map[string]any
				json.Unmarshal([]byte(toolCall.Arguments), &args)
				location := args["location"].(string)

				// Simulate getting weather data
				// weatherData := getWeather(location)
				fmt.Printf("Weather in %s\n", location)

				// Continue conversation with function result
				// response, _ = client.Responses.New(ctx, responses.ResponseNewParams{
				// 	Model:              openai.ChatModelGPT5_2,
				// 	PreviousResponseID: openai.String(response.ID),
				// 	Input: responses.ResponseNewParamsInputUnion{
				// 		OfInputItemList: []responses.ResponseInputItemUnionParam{{
				// 			OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
				// 				CallID: toolCall.CallID,
				// 				Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
				// 					OfString: openai.String(weatherData),
				// 				},
				// 			},
				// 		}},
				// 	},
				// })
			}
		}
	}

	// for debugging
	// fmt.Println("RAW:", raw)

	// dec := json.NewDecoder(strings.NewReader(raw))

	var agentResp AgentResponse
	// if err := dec.Decode(&agentResp); err != nil {
	// 	return AgentResponse{}, err
	// }

	return agentResp, nil
}

func buildPrompt(messages []Message) string {
	content, err := prompts.Files.ReadFile("system.txt")
	if err != nil {
		panic(err)
	}

	prompt := string(content) + "\n\nConversation:\n"

	for _, m := range messages {
		prompt += m.Role + ": " + m.Content + "\n"
	}

	return prompt
}
