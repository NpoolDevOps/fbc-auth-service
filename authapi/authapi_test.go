package authapi

import (
	types "github.com/NpoolDevOps/fbc-auth-service/types"
	"github.com/google/uuid"
	"testing"
)

func TestLogin(t *testing.T) {
	Login(types.UserLoginInput{
		Username: "entropypool",
		Password: "b759803bc603",
		AppId:    uuid.MustParse("00000000-0000-0000-0000-000000000000"),
	})
}

func TestCheckSuperUser(t *testing.T) {
	CheckSuperUser(types.CheckSuperUserInput{
		AuthCode: "9a084e991104f774f1a8e56c30af6f4abd9696c24757835e0d5c4991ba122f8c",
	})
}
