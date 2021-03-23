package types

import (
	"github.com/google/uuid"
)

type UserLoginInput struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	AppId     uuid.UUID `json:"appid"`
	TargetUrl string    `json:"target_url"`
}

type UserLoginOutput struct {
	AuthCode  string `json:"auth_code"`
	TargetUrl string `json:"target_url"`
}

type UserInfoInput struct {
	AuthCode string `json:"auth_code"`
}

type UserInfoOutput struct {
	Id          uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	SuperUser   bool      `json:"super_user"`
	VisitorOnly bool      `json:"visitor_only"`
}

type ModifyPasswordInput struct {
	AuthCode    string `json:"auth_code"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
