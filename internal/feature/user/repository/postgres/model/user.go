package model

// In models, you can write DB tags only.
// It is optional, because usually struct field will be interpreted in snake_case.
type User struct {
	ID    uint64
	Email string
}

type UserInfo struct {
	ID        uint64
	Firstname string
	Lastname  string
	UserID    uint64
}
