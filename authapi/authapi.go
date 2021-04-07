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

func Login(input types.UserLoginInput) (*types.UserLoginOutput, error) {
	host, err := etcdcli.GetHostByDomain(authDomain)
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
	host, err := etcdcli.GetHostByDomain(authDomain)
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

func UsernameInfo(input types.UsernameInfoInput) (*types.UsernameInfoOutput, error) {
	host, err := etcdcli.GetHostByDomain(authDomain)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get %v from etcd: %v", authDomain, err)
		return nil, err
	}

	log.Infof(log.Fields{}, "req to http://%v%v", host, types.UsernameInfoAPI)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v%v", host, types.UsernameInfoAPI))
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

	output := types.UsernameInfoOutput{}
	b, _ := json.Marshal(apiResp.Body)
	err = json.Unmarshal(b, &output)

	return &output, err
}

func VisitorOwner(input types.VisitorOwnerInput) (*types.VisitorOwnerOutput, error) {
	host, err := etcdcli.GetHostByDomain(authDomain)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to get %v from etcd: %v", authDomain, err)
		return nil, err
	}

	log.Infof(log.Fields{}, "req to http://%v%v", host, types.VisitorOwnerAPI)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v%v", host, types.VisitorOwnerAPI))
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

	output := types.VisitorOwnerOutput{}
	b, _ := json.Marshal(apiResp.Body)
	err = json.Unmarshal(b, &output)

	return &output, err
}
