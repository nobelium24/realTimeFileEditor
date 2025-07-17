package model

type Role string

const (
	Edit    Role = "edit"
	Read    Role = "read"
	Creator Role = "creator"
)

type InviteStatus string

const (
	Pending  Role = "pending"
	Accepted Role = "accepted"
	Declined Role = "declined"
)
