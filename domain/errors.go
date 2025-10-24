package domain

// Application errors
var (
	ErrUserNotFound    = &AppError{"user not found", 404}
	ErrChatNotFound    = &AppError{"chat not found", 404}
	ErrMessageNotFound = &AppError{"message not found", 404}
	ErrUsernameExists  = &AppError{"username already exists", 409}
	ErrInvalidUser     = &AppError{"invalid user", 400}
	ErrEmptyMessage    = &AppError{"message content cannot be empty", 400}
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Message string
	Code    int
}

func (e *AppError) Error() string {
	return e.Message
}
