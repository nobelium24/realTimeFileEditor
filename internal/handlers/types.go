package handlers

type PasswordResetCode struct {
	FullName  string
	ResetCode string
	Year      int
}

type WelcomeMessage struct {
	FullName string
	Year     int
}
