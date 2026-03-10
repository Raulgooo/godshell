package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	ctxengine "godshell/context"
	"godshell/llm"
	"godshell/observer"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== Godshell AI Debugging REPL ===")

	// 1. Prompt for OpenRouter Key
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter OpenRouter API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		log.Fatal("API key is required")
	}

	// 2. Initialize Daemon
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	tree := ctxengine.NewProcessTree()
	go tree.EvictGhosts(60 * time.Second)
	go tree.RefreshMetrics(5 * time.Second)

	events := make(chan observer.Event, 256)
	go func() {
		if err := observer.Run(ctx, events); err != nil {
			log.Printf("Observer error: %v", err)
		}
	}()

	go func() {
		for {
			select {
			case e := <-events:
				tree.HandleEvent(e)
			case <-ctx.Done():
				return
			}
		}
	}()

	// 3. Setup LLM Client (OpenRouter endpoint)
	// Using google/gemini-2.0-flash-001 or similar large context model
	client := llm.NewOpenAIClient(apiKey, "https://openrouter.ai/api/v1/chat/completions", "google/gemini-2.0-flash-001")

	fmt.Println("\nDaemon started. Waiting for initial events (5s)...")
	time.Sleep(5 * time.Second)

	// 4. Start Conversational Loop
	snap := tree.TakeSnapshot()
	conv := llm.NewConversation(snap)

	fmt.Println("\nReady. Type your question (or 'refresh' to update snapshot, 'exit' to quit):")
	for {
		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}
		if input == "refresh" {
			snap = tree.TakeSnapshot()
			conv.UpdateSnapshot(snap)
			fmt.Printf("Snapshot refreshed at %s\n", snap.Timestamp.Format(time.Kitchen))
			continue
		}
		if input == "" {
			continue
		}

		// Add User message
		conv.History = append(conv.History, llm.Message{
			Role:    llm.RoleUser,
			Content: input,
		})

		// Conversation loop (handles potential chaining of tool calls)
		for {
			resp, err := client.Chat(conv.History, conv.GetToolDefinitions())
			if err != nil {
				fmt.Printf("LLM Error: %v\n", err)
				break
			}

			// Add Assistant message to history
			conv.History = append(conv.History, *resp)

			if len(resp.ToolCalls) > 0 {
				fmt.Printf("[%d tool calls received]\n", len(resp.ToolCalls))
				for _, tc := range resp.ToolCalls {
					fmt.Printf("  - Tool: %s\n", tc.Function.Name)

					var args map[string]interface{}
					json.Unmarshal([]byte(tc.Function.Arguments), &args)

					result, err := conv.ExecuteTool(tc.Function.Name, args)
					if err != nil {
						result = fmt.Sprintf("Error: %v", err)
					}

					// Append tool response
					conv.History = append(conv.History, llm.Message{
						Role:       llm.RoleTool,
						ToolCallID: tc.ID,
						Name:       tc.Function.Name,
						Content:    result,
					})
				}
				// LLM gets the tool results in the next loop iteration
				continue
			}

			// Final natural language response
			fmt.Printf("\nAI: %s\n", resp.Content)
			break
		}
	}

	fmt.Println("Exiting.")
}
