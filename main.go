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

type ApiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Body interface{} `json:"body"`
}

var srvs []SrvInfo
var cli *resty.Client

func init() {
	discoverService(MyConfig.Domain)
	cli = resty.New()
}

func parseResponseBody(resBody []byte) *ApiResp {
	var unmar map[string]interface{}
	_ = json.Unmarshal(resBody, &unmar)

	parseRes := new(ApiResp)
	parseRes.Code = int(unmar["code"].(float64))
	parseRes.Msg = unmar["msg"].(string)
	parseRes.Body = unmar["body"]

	return parseRes
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

	apiResp := parseResponseBody(resp.Body())

	if apiResp.Code == 0 {
		body := apiResp.Body.(map[string]interface{})
		jsonBody, _ := json.Marshal(map[string]string{"auth_code": body["auth_code"].(string), "url": jumpUrl})
		_ = json.Unmarshal(jsonBody, &apiResp.Body)
	}

	if apiResp.Code == 5 {
		elog.Errorf(elog.Fields{}, "%v", apiResp.Msg)
		apiResp.Code = 2
		apiResp.Msg = "server exception"
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

	fmt.Println(resp, err)

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

func discoverService(serviceKey string) {
	ips, err := srvreg.Query(serviceKey)
	if err != nil {
		elog.Fatalf(elog.Fields{}, "discover service failed: %v", serviceKey)
	}
	var srvInfo SrvInfo
	srvInfo.Domain = serviceKey
	srvInfo.IpPorts = ips

	srvs = append(srvs, srvInfo)
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
