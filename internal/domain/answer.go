package domain

type AnswerResult struct {
	Answer  string         `json:"answer"`
	Sources []SearchResult `json:"sources"`
}
