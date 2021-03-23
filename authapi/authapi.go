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
		log.Errorf(log.Fields{}, "cannot get %v: %v", authDomain, err)
		return "", err
	}

	err = json.Unmarshal([]byte(resp[0]), &myConfig)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse %v: %v", string(resp[0]), err)
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

	log.Infof(log.Fields{}, "req to http://%v%v", host, types.UserLoginAPI)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v%v", host, types.UserLoginAPI))
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
	b, _ := json.Marshal(apiResp.Body)
	err = json.Unmarshal(b, &output)

	return &output, err
}

func UserInfo(input types.UserInfoInput) (*types.UserInfoOutput, error) {
	host, err := getAuthHost()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get %v from etcd: %v", authDomain, err)
		return nil, err
	}

	log.Infof(log.Fields{}, "req to http://%v%v", host, types.UserInfoAPI)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v%v", host, types.UserInfoAPI))
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

	output := types.UserInfoOutput{}
	b, _ := json.Marshal(apiResp.Body)
	err = json.Unmarshal(b, &output)

	return &output, err
}

func CheckUser(input types.CheckUserInput) (bool, error) {
	host, err := getAuthHost()
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get %v from etcd: %v", authDomain, err)
		return false, err
	}

	log.Infof(log.Fields{}, "req to http://%v%v", host, types.CheckUserAPI)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v%v", host, types.CheckUserAPI))
	if err != nil {
		log.Errorf(log.Fields{}, "heartbeat error: %v", err)
		return false, err
	}

	if resp.StatusCode() != 200 {
		return false, xerrors.Errorf("NON-200 return")
	}

	_, err = httpdaemon.ParseResponse(resp)
	if err != nil {
		return false, err
	}

	return true, err
}
