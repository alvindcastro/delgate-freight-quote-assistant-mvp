package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
)

type Config struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

type Service struct {
	config Config
	client *http.Client
}

type Output struct {
	Summary           string   `json:"summary"`
	CustomerMessage   string   `json:"customerMessage"`
	FollowUpQuestions []string `json:"followUpQuestions"`
	Provider          string   `json:"provider"`
}

func NewService(config Config) *Service {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1/chat/completions"
	}
	return &Service{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

func (s *Service) Generate(ctx context.Context, req quote.Request, resp quote.Response) Output {
	if strings.TrimSpace(s.config.APIKey) != "" {
		if output, err := s.generateWithOpenAI(ctx, req, resp); err == nil {
			output.Provider = "openai"
			return output
		}
	}

	output := Local(req, resp)
	output.Provider = "local-fallback"
	return output
}

func Local(req quote.Request, resp quote.Response) Output {
	summary := fmt.Sprintf(
		"%s quote estimate for %s is %s %.0f-%.0f. Chargeable weight is %.0f lbs. Status: %s. Confidence: %s.",
		humanizeService(req.ServiceLevel),
		resp.RouteLabel,
		resp.Currency,
		resp.EstimatedLow,
		resp.EstimatedHigh,
		resp.ChargeableWeightLbs,
		resp.Status,
		resp.Confidence,
	)

	if len(resp.MissingFields) > 0 {
		summary += " Missing details should be collected before sending a final quote: " + strings.Join(resp.MissingFields, ", ") + "."
	}
	if len(resp.RiskFlags) > 0 {
		summary += " Operations should review: " + strings.Join(resp.RiskFlags, "; ") + "."
	}
	if resp.Breakdown.AccessorialFees > 0 {
		summary += fmt.Sprintf(" Accessorial charges are included at %s %.0f.", resp.Currency, resp.Breakdown.AccessorialFees)
	}

	customerName := strings.TrimSpace(req.CustomerName)
	greeting := "Hello,"
	if customerName != "" {
		greeting = "Hello " + customerName + ","
	}

	customerMessage := fmt.Sprintf(
		"%s\n\nBased on the shipment details provided, the estimated freight range for %s is %s %.0f-%.0f. This estimate includes the selected service level and any requested accessorial services.\n\nPlease note this is a demo estimate and not a binding quote. Final pricing should be confirmed after pickup and delivery details, carrier availability, and operational requirements are reviewed.",
		greeting,
		resp.RouteLabel,
		resp.Currency,
		resp.EstimatedLow,
		resp.EstimatedHigh,
	)

	questions := followUpQuestions(resp)
	if len(questions) > 0 {
		customerMessage += "\n\nTo finalize the quote, please confirm:\n- " + strings.Join(questions, "\n- ")
	}

	return Output{
		Summary:           summary,
		CustomerMessage:   customerMessage,
		FollowUpQuestions: questions,
	}
}

func (s *Service) generateWithOpenAI(ctx context.Context, req quote.Request, resp quote.Response) (Output, error) {
	type chatRequest struct {
		Model       string        `json:"model"`
		Messages    []chatMessage `json:"messages"`
		Temperature float64       `json:"temperature"`
	}
	type chatResponse struct {
		Choices []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}

	promptData := map[string]any{
		"request":  req,
		"response": resp,
	}
	payloadBytes, _ := json.MarshalIndent(promptData, "", "  ")

	body := chatRequest{
		Model: s.config.Model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are a concise logistics operations assistant. Return only valid JSON with keys summary, customerMessage, and followUpQuestions. Do not change the quote price. Do not present the estimate as binding."},
			{Role: "user", Content: "Create an operations summary and customer-ready message for this freight quote. Keep it professional and practical. JSON input:\n" + string(payloadBytes)},
		},
		Temperature: 0.2,
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return Output{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.BaseURL, bytes.NewReader(encoded))
	if err != nil {
		return Output{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return Output{}, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return Output{}, fmt.Errorf("OpenAI request failed with status %d", httpResp.StatusCode)
	}

	var decoded chatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&decoded); err != nil {
		return Output{}, err
	}
	if len(decoded.Choices) == 0 {
		return Output{}, fmt.Errorf("OpenAI response contained no choices")
	}

	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var modelOutput openAIOutput
	if err := json.Unmarshal([]byte(content), &modelOutput); err != nil {
		return Output{}, err
	}
	output, err := validateOpenAIOutput(modelOutput, resp)
	if err != nil {
		return Output{}, err
	}
	return output, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIOutput struct {
	Summary           *string   `json:"summary"`
	CustomerMessage   *string   `json:"customerMessage"`
	FollowUpQuestions *[]string `json:"followUpQuestions"`
}

func validateOpenAIOutput(modelOutput openAIOutput, resp quote.Response) (Output, error) {
	if modelOutput.Summary == nil {
		return Output{}, fmt.Errorf("OpenAI output missing summary")
	}
	if modelOutput.CustomerMessage == nil {
		return Output{}, fmt.Errorf("OpenAI output missing customerMessage")
	}
	if modelOutput.FollowUpQuestions == nil {
		return Output{}, fmt.Errorf("OpenAI output missing followUpQuestions")
	}

	summary := strings.TrimSpace(*modelOutput.Summary)
	customerMessage := strings.TrimSpace(*modelOutput.CustomerMessage)
	questions := *modelOutput.FollowUpQuestions

	if summary == "" {
		return Output{}, fmt.Errorf("OpenAI output summary is empty")
	}
	if customerMessage == "" {
		return Output{}, fmt.Errorf("OpenAI output customerMessage is empty")
	}
	if !containsQuoteFacts(summary, resp, true) {
		return Output{}, fmt.Errorf("OpenAI summary changed quote facts")
	}
	if !containsQuoteFacts(customerMessage, resp, false) {
		return Output{}, fmt.Errorf("OpenAI customerMessage changed quote facts")
	}
	if !containsNonBindingDisclaimer(customerMessage) {
		return Output{}, fmt.Errorf("OpenAI customerMessage missing non-binding disclaimer")
	}
	if !sameStrings(questions, followUpQuestions(resp)) {
		return Output{}, fmt.Errorf("OpenAI followUpQuestions changed required follow-up questions")
	}

	return Output{
		Summary:           summary,
		CustomerMessage:   customerMessage,
		FollowUpQuestions: questions,
	}, nil
}

func containsQuoteFacts(text string, resp quote.Response, requireInternalFacts bool) bool {
	if !strings.Contains(text, resp.RouteLabel) {
		return false
	}
	if !strings.Contains(text, estimatedRange(resp)) {
		return false
	}
	if requireInternalFacts {
		if !strings.Contains(text, fmt.Sprintf("Chargeable weight is %.0f lbs", resp.ChargeableWeightLbs)) {
			return false
		}
		if !strings.Contains(text, "Status: "+resp.Status) {
			return false
		}
		if !strings.Contains(text, "Confidence: "+resp.Confidence) {
			return false
		}
	}
	return true
}

func estimatedRange(resp quote.Response) string {
	return fmt.Sprintf("%s %.0f-%.0f", resp.Currency, resp.EstimatedLow, resp.EstimatedHigh)
}

func containsNonBindingDisclaimer(text string) bool {
	normalized := strings.ToLower(text)
	return (strings.Contains(normalized, "not a binding") || strings.Contains(normalized, "non-binding")) &&
		strings.Contains(normalized, "final pricing")
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func followUpQuestions(resp quote.Response) []string {
	questions := make([]string, 0)
	for _, field := range resp.MissingFields {
		switch field {
		case "Pickup postal code":
			questions = append(questions, "What is the pickup postal code?")
		case "Delivery postal code":
			questions = append(questions, "What is the delivery postal code?")
		case "Shipment weight":
			questions = append(questions, "What is the total shipment weight?")
		case "Shipment dimensions":
			questions = append(questions, "What are the shipment dimensions?")
		case "Piece or pallet count":
			questions = append(questions, "How many pieces or pallets are included?")
		default:
			questions = append(questions, "Please confirm: "+strings.ToLower(field)+".")
		}
	}
	return questions
}

func humanizeService(service string) string {
	switch strings.ReplaceAll(strings.ToLower(service), "_", " ") {
	case "same day", "same-day", "sameday":
		return "Same-day"
	case "expedited", "rush":
		return "Expedited"
	default:
		return "Standard"
	}
}
