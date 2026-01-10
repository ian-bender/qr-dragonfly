package httpapi

import "user-service/internal/model"

type AuthSession struct {
	User  model.User `json:"user"`
	Token string     `json:"token,omitempty"`
}

type createUserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	UserType string `json:"userType,omitempty"`
}

type updateUserInput struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	Name     *string `json:"name,omitempty"`
	UserType *string `json:"userType,omitempty"`

	// Optional management
	Disabled *bool `json:"disabled,omitempty"`
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type confirmSignUpInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type resendConfirmationInput struct {
	Email string `json:"email"`
}

type forgotPasswordInput struct {
	Email string `json:"email"`
}

type confirmForgotPasswordInput struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type changePasswordInput struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}
