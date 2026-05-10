package agent

import (
	"bufio"
	"fmt"
	"os"
	"simple-k8s-agent/internal/tools"
	"strings"
)

type Message struct {
	Role    string
	Content string
}

type ToolFunc func(input map[string]string) (string, error)

type Agent struct {
	brain    LLM
	messages []Message
	tools    map[string]ToolFunc
}

func New(llm LLM) *Agent {
	return &Agent{
		brain:    llm,
		messages: []Message{},
		tools: map[string]ToolFunc{
			"getContext": func(input map[string]string) (string, error) {
				return tools.GetContexts(input)
			},
			"setContext": func(input map[string]string) (string, error) {
				return tools.SetContexts(input)
			},
			"getPods": func(input map[string]string) (string, error) {
				return tools.GetPods(input)
			},
			"getNamespaces": func(input map[string]string) (string, error) {
				return tools.GetNamespaces(input)
			},
			"describePod": func(input map[string]string) (string, error) {
				return tools.DescribePod(input)
			},
			"getPodLogs": func(input map[string]string) (string, error) {
				return tools.GetPodLogs(input)
			},
		},
	}
}

func (a *Agent) Run(input string) error {
	a.messages = append(a.messages, Message{
		Role:    "user",
		Content: input,
	})
	steps := 0
	for {
		steps++
		if steps > 10 {
			return fmt.Errorf("agent stopped: max num of iterration rached (%d)", steps)
		}
		resp, err := a.brain.Call(a.messages)
		if err != nil {
			return err
		}

		switch resp.Type {

		case "ask":
			fmt.Println("Agent:", resp.Question)

			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Input: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return err
			}

			input = strings.TrimSpace(input)

			a.messages = append(a.messages, Message{
				Role:    "user",
				Content: input,
			})
		case "tool":
			fmt.Println("Calling tool:", resp.Tool)
			tool, ok := a.tools[resp.Tool]
			if !ok {
				return fmt.Errorf("unknown tool: %s", resp.Tool)
			}

			result, err := tool(resp.Input)
			if err != nil {
				return err
			}

			a.messages = append(a.messages, Message{
				Role: "tool",
				Content: fmt.Sprintf(
					"tool=%s input=%v result=%s",
					resp.Tool,
					resp.Input,
					result,
				),
			})

		case "final":
			fmt.Println(resp.Answer)
			return nil
		}
	}
}
