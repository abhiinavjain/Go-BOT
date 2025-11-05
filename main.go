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

func loadFaqs(filepath string) (string, error) {
	filedata, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("Could not read file")

	}

	return string(filedata), nil
}

func botprompt(faqs string) string {
	outofscope := "This question is not related to customer support, you can exit this chat by typing (exit) "

	return fmt.Sprintf(`You are a friendly customer support bot for 'My Company'.
		 Your name is 'ResolveAl'. You must follow these rules in this exact order:
   		 Check FAQs: First, ALWAYS check if the user's question can be answered by the 'Common FAQs' data provided below. 
		 If the user's question is a semantic match (e.g., "when are you open?" is a match for "What are your business hours?"), 
		 you MUST provide the answer from the matching FAQ. 
		 [Common FAQs Start] %s [Common FAQs End]
		 Check for Customer Support Topic: If the question is NOT in the FAQs, 
		 check if it is a general customer support-related question (e.Example, about billing, shipping, user accounts, troubleshooting, 
		 refunds, etc.). If it IS, answer it helpfully using your general knowledge.
		Handle Out-of-Scope: If the question is NOT in the FAQs and is NOT related to customer support (e.g., 'what is the capital of France?', 'tell me a joke', 'who won the game?'), 
		you MUST respond with EXACTLY this message: %q
		Keep your answers concise and professional. `, faqs, outofscope)
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

	faqContent, _ := loadFaqs("faqs.json")

	systemPrompt := botprompt(faqContent)
	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}

	fmt.Println("FAQs Loaded Successfully!")

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
		resp, err := chat.SendMessage(ctx, genai.Text(promnt))

		if err != nil {
			fmt.Printf("Bot: Sorry, I had an error: %v\n", err)
			continue
		}

		printResponse(resp)

	}
}
