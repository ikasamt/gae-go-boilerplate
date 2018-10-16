package app

import (
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
)

type PasswordReset struct {
	ID         int
	UserID     int
	Token      string
	Status     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	beforeJSON gin.H
	errors     error
}

func (x PasswordReset) User() User {
	return fetchUser(x.UserID)
}

// 制約条件
func (x PasswordReset) Validations() error {
	return validation.ValidateStruct(&x,
		validation.Field(&x.UserID, validation.Required),
	)
}
