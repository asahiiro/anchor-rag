package generation

import "context"

type Role string

const (
	RoleSystem Role = "system"
	RoleUser   Role = "user"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type Generator interface {
	Generate(
		ctx context.Context,
		messages []Message,
	) (string, error)
}
