package data

// Role represents the role of a user in the system.
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleTeacher Role = "teacher"
)