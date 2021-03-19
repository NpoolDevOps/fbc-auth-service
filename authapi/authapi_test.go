package authapi

import (
	types "github.com/NpoolDevOps/fbc-auth-service/types"
	"github.com/google/uuid"
	"testing"
)

func TestLogin(t *testing.T) {
	Login(types.UserLoginInput{
		Username: "entropypool",
		Password: "7d1721d7acef",
		AppId:    uuid.MustParse("00000000-0000-0000-0000-000000000000"),
	})
}
