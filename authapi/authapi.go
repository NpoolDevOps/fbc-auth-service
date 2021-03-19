package authapi

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	types "github.com/NpoolDevOps/fbc-auth-service/types"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
)

const authDomain = "auth.npool.top"

type authHostConfig struct {
	Host string `json:"host"`
}

func getAuthHost() (string, error) {
	var myConfig authHostConfig

	resp, err := etcdcli.Get(authDomain)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(resp[0], &myConfig)
	if err == nil {
		return "", err
	}

	return myConfig.Host, err
}

func Login(input types.UserLoginInput) (*types.UserLoginOutput, error) {
	host, err := getAuthHost()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get %v from etcd: %v", authDomain, err)
		return nil, err
	}

	log.Infof(log.Fields{}, "req to http://%v/%v", host, types.UserLoginAPI)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v/%v", host, types.UserLoginAPI))
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
