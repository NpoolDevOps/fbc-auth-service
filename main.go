package main

import (
	"encoding/json"
	"fmt"
	elog "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolRD/http-daemon"
	"github.com/NpoolRD/service-register"
	"github.com/go-resty/resty/v2"
	"net/http"
)

type SrvInfo struct {
	Domain  string   `json:"domain"`
	IpPorts []string `json:"ip_ports"`
}

type LoginRequest struct {
	Username string
	Passwd   string
	AppId    string
}

var srvs []srvreg.SrvInfo
var cli *resty.Client

func init() {
	srvs = srvreg.BatchQuery(MyConfig.Targets)
	cli = resty.New()
}

func serveLogin(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	params := req.Form
	if err := httpdaemon.ValidateParams([]string{"username", "passwd"}, params); err != nil {
		return nil, err.Error(), 1
	}
	username := params["username"][0]
	passwd := params["passwd"][0]

	var jumpUrl = ""
	if len(params["url"]) > 0 {
		jumpUrl = params["url"][0]
	}

	var appId = ""
	url := getUrl("/auth/login")

	resp, err := cli.R().
		SetQueryParam("username", username).
		SetQueryParam("passwd", passwd).
		SetQueryParam("appid", appId).
		Post(url)

	if err != nil {
		elog.Errorf(elog.Fields{}, "require failed: %v", err)
		return nil, "server exception", 2
	}

	apiResp := httpdaemon.ParseResponseBody(resp.Body())

	if apiResp.Code == 0 {
		body := apiResp.Body.(map[string]interface{})
		jsonBody, _ := json.Marshal(map[string]string{"auth_code": body["auth_code"].(string), "url": jumpUrl})
		_ = json.Unmarshal(jsonBody, &apiResp.Body)
	}

	if apiResp.Code != 0 {
		elog.Errorf(elog.Fields{}, "login failed %v", apiResp.Msg)
		apiResp.Msg = fmt.Sprintf("login failed code %v , error msg %s", apiResp.Code, apiResp.Msg)
		apiResp.Code = 2
	}

	return apiResp.Body, apiResp.Msg, apiResp.Code
}

func serveLogout(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	params := req.Form
	if err := httpdaemon.ValidateParams([]string{"auth_code"}, params); err != nil {
		return nil, err.Error(), 1
	}
	authCode := params["auth_code"][0]
	url := getUrl("/auth/logout")

	resp, err := cli.R().
		SetQueryParam("auth_code", authCode).
		Post(url)

	if err != nil {
		elog.Errorf(elog.Fields{}, "require failed: %v", err)
		return nil, "server exception", 2
	}

	return nil, "", 0
}

func getUrl(uri string) string {
	url := "http://" + srvs[0].IpPorts[0]

	return url + uri
}

func main() {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/login",
		Handler:  serveLogin,
		Method:   "POST",
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/logout",
		Handler:  serveLogout,
		Method:   "POST",
	})

	httpdaemon.Run(MyConfig.Port)

	ch := make(chan int, 0)
	<-ch
}
