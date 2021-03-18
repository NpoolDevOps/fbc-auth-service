package types

import (
	"github.com/google/uuid"
)

type UserLoginInput struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	AppId    uuid.UUID `json:"appid"`
}

type UserLoginOutput struct {
	AuthCode string `json:"auth_code"`
}
