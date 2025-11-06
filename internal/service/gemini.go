package service

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"
)

const (
	outofscope = "This question is not related to customer support."
	filepath   = "faq.json"
)

type GeminiService struct {
	model *genai.GenerativeModel
	db    *pgxpool.Pool
}

func BotService(ctx context.Context, apikey string, dbUrl string) (*GeminiService, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apikey))

	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	dbpool, err := pgxpool.New(ctx, dbUrl)

	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	model := client.GenerativeModel("models/gemini-flash-latest")

	faqContent, err := loadFaqs(filepath)

	if err != nil {
		return nil, err
	}

	systemPrompt := botprompt(faqContent)

	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}

	return &GeminiService{
		model: model,
		db:    dbpool,
	}, nil
}

func (g *GeminiService) CreateNewSession(ctx context.Context) (string, error) {
	sql_query := "INSERT INTO chat_sessions DEFAULT VALUES RETURNING id"
	var sessionID string
	err := g.db.QueryRow(ctx, sql_query).Scan(&sessionID)

	if err != nil {
		return "", err
	}

	return sessionID, nil

}

func (g *GeminiService) AddMessageToHistory(ctx context.Context, sessionID string, role string, content string) error {
	sql_query := "INSERT INTO chat_messages (session_id, role, content) VALUES ($1, $2, $3)"
	_, err := g.db.Exec(ctx, sql_query, sessionID, role, content)
	if err != nil {
		return err
	}

	return nil

}

func (g *GeminiService) GetChatHistory(ctx context.Context, sessionId string) ([]*genai.Content, error) {
	sql := "SELECT role, content FROM chat_messages WHERE session_id = $1 ORDER BY created_at ASC"
	var history []*genai.Content
	rows, err := g.db.Query(ctx, sql, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var role, content string

		if err := rows.Scan(&role, &content); err != nil {
			return nil, err
		}
		msq := &genai.Content{Role: role, Parts: []genai.Part{genai.Text(content)}}
		history = append(history, msq)
	}

	return history, nil
}

func loadFaqs(filepath string) (string, error) {
	filedata, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("Could not read file")

	}

	return string(filedata), nil
}

func botprompt(faqs string) string {

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

func (s *GeminiService) RespGenerator(ctx context.Context, sessionID string, promt string) (string, error) {
	history, err := s.GetChatHistory(ctx, sessionID)
	if err != nil {
		return "", err
	}

	err = s.AddMessageToHistory(ctx, sessionID, "user", promt)

	if err != nil {
		return "", err
	}

	chat := s.model.StartChat()

	chat.History = history

	resp, err := chat.SendMessage(ctx, genai.Text(promt))

	if err != nil {
		return "", fmt.Errorf("%v", err)
	}

	var botresp string

	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				botresp = string(txt)
				break
			}
		}
	}

	err = s.AddMessageToHistory(ctx, sessionID, "model", botresp)

	if err != nil {
		return "", err
	}

	return botresp, err

}
