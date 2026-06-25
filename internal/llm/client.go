package llm

import "context"

// Provider defines the interface for LLM classification/draft.
type Provider interface {
	Classify(ctx context.Context, in ClassifyInput) (ClassifyResult, error)
	DraftReply(ctx context.Context, in DraftInput) (string, error)
}

// ClassifyInput holds email metadata for LLM analysis.
type ClassifyInput struct {
	Subject    string
	Body       string
	Sender     string
	Candidates []string
}

// ClassifyResult is the structured output from LLM analysis.
type ClassifyResult struct {
	Summary  string   `json:"summary"`
	Priority int      `json:"priority"`
	Labels   []string `json:"labels"`
	IsSpam   bool     `json:"is_spam"`
}

// DraftInput holds context for generating a reply draft.
type DraftInput struct {
	ThreadText string
	UserHint   string
}

// OpenAIProvider implements Provider using an OpenAI-compatible endpoint.
type OpenAIProvider struct{}

var _ Provider = (*OpenAIProvider)(nil)

func NewOpenAIProvider(baseURL, apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{}
}

func (p *OpenAIProvider) Classify(ctx context.Context, in ClassifyInput) (ClassifyResult, error) {
	return ClassifyResult{}, nil
}

func (p *OpenAIProvider) DraftReply(ctx context.Context, in DraftInput) (string, error) {
	return "", nil
}
