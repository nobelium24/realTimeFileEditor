package handlers

import "realTimeEditor/internal/model"

type PasswordResetCode struct {
	FullName  string
	ResetCode string
	Year      int
}

type WelcomeMessage struct {
	FullName string
	Year     int
}

type Invite struct {
	FullName      string
	DocumentTitle string
	Role          model.Role
	InviteLink    string
	Year          int
}

type AccountSetup struct {
	DocumentTitle    string
	Role             model.Role
	AccountSetupLink string
}
