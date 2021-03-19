package authapi

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-auth-service/types"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
)

func Login(input types.UserLoginInput) (*types.UserLoginOutput, error) {
	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(types.UserLoginAPI)
	if err != nil {
		log.Errorf(log.Fields{}, "heartbeat error: %v", err)
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, xerrors.Errorf("NON-200 return")
	}

	apiResp, err := httpdaemon.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	output := types.UserLoginOutput{}
	err = json.Unmarshal([]byte(apiResp.Body.(string)), &output)

	return &output, err
}
