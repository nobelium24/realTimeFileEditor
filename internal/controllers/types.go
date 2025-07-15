package controllers

type LoginPayload struct {
	Email    *string `json:"email"`
	UserName *string `json:"userName"`
	Password string  `json:"password"`
}
