package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				fmt.Println(txt)
			}
		}
	}
}

func main() {
	_ = godotenv.Load()

	apiKey := os.Getenv("api_key")
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))

	if err != nil {
		log.Fatal("Failed to create api client", err)
	}

	defer client.Close()

	model := client.GenerativeModel("models/gemini-flash-latest")
	chat := model.StartChat()
	chat.History = []*genai.Content{}

	fmt.Printf("Hello how can I help you ? ,\n")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		promnt := strings.TrimSpace(input)
		if strings.ToLower(promnt) == "quit" || strings.ToLower(promnt) == "exit" {
			fmt.Println("Bot: Goodbye!")
			break
		}
		fmt.Println("Sending prompt: ", promnt)
		fmt.Println("------")
		resp, err := chat.SendMessage(ctx, genai.Text(promnt))

		if err != nil {
			fmt.Printf("Bot: Sorry, I had an error: %v\n", err)
			continue
		}

		printResponse(resp)

	}
}
