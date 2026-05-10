package main

import (
	"fmt"
	"os"
	"simple-k8s-agent/internal/agent"
)

func main() {
	// read user input
	if len(os.Args) > 2 {
		fmt.Println("Usage: simple-k8s-agent 'why my pod A is not running?'")
		os.Exit(1)
	}

	userInput := "Example: Is there any app which is not running well in default namespace?"
	// userInput := os.Args[1]
	fmt.Printf("Input: %s\n", userInput)

	llm := agent.NewOpenAIBrain()
	a := agent.New(llm)

	if err := a.Run(userInput); err != nil {
		os.Exit(1)
	}
}
