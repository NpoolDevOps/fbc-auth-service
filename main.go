package main

import (
	"fmt"
	"net/http"
	elog "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolRD/http-daemon"
	"github.com/NpoolRD/service-register"
	"github.com/go-resty/resty/v2"
	"strings"
	"io/ioutil"
	"encoding/json"
	nurl "net/url"
)

type SrvInfo struct {
	Domain string `json:"domain"`
	IpPorts []string `json:"ip_ports"`
}

type LoginRequest struct {
	Username string
	Passwd string
	AppId string
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

func Post(url string, data map[string]string) (interface{}, error) {
	values := makeValues(data)
	info, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
	if err != nil {
		elog.Errorf(elog.Fields{}, "http post failed %v", err)
		return nil, err
	}
	defer info.Body.Close()
	body, err := ioutil.ReadAll(info.Body)

	if err != nil {
		elog.Errorf(elog.Fields{}, "read response body failed %v", err)
		return nil, err
	}

	var resp interface{}
	err = json.Unmarshal(body, &resp)

	if err != nil {
		elog.Errorf(elog.Fields{}, "json unmarshal failed %v", err)
		return nil, err
	}

	return resp, nil
}

func makeValues(data map[string]string) *nurl.Values {
	params := nurl.Values{}
	for key, val := range data {
		params.Set(key, val)
	}

	return &params
}

func serveLogin(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	params := req.Form
	if err := httpdaemon.ValidateParams([]string{"username", "passwd"}, params); err != nil {
		return nil, err.Error(), 1
	}
	username := params["username"][0]
	passwd := params["passwd"][0]
	/*
	var jumpUrl = ""
	if len(params["url"]) > 0 {
		jumpUrl = params["url"][0]
	}
	*/
	//var appId = ""
	url := getUrl("/auth/login")

	/**
	resp, err := cli.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{"username": username, "passwd": passwd, "appid": appId}).
	//	SetResult(&AuthSuccess{}).
		Post(url)

	fmt.Println("lll", resp, url, jumpUrl, err)
	*/

	resp, err := Post(url, map[string]string{"username":username, "passwd":passwd})
	if err != nil {
		elog.Errorf(elog.Fields{}, "require failed: %v", err)
		return nil, "server exception", 2
	}

	postRes := resp.(map[string]interface{})
	code := int(postRes["code"].(float64))
	msg := postRes["msg"].(string)
	if code == 5 {
		elog.Errorf(elog.Fields{}, "%v", msg)
		code = 2
		msg = "server exception"
	}

	body := postRes["body"].(map[string]interface{})
	jsonBody, _ := json.Marshal(map[string]string{"auth_code": body["auth_code"].(string)})
	var respo interface{}
	_ = json.Unmarshal(jsonBody, &respo)

	return respo, msg, code
}

func serveLogout(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	params := req.Form
	if err := httpdaemon.ValidateParams([]string{"auth_code"}, params); err != nil {
		return nil, err.Error(), 1
	}
	authCode := params["auth_code"][0]
	url := getUrl("/auth/logout")
	resp, err := Post(url, map[string]string{"auth_code":authCode})
	if err != nil {
		elog.Errorf(elog.Fields{}, "require failed: %v", err)
		return nil, "server exception", 2
	}
	fmt.Println(resp)

	return nil, "", 0
}

func getUrl(uri string) string {
	url := "http://106.14.125.55:40001"
	//url := "http://localhost:40001"

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
		Method: "POST",
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/logout",
		Handler:  serveLogout,
		Method: "POST",
	})

	httpdaemon.Run(MyConfig.Port)

	ch := make(chan int, 0)
	<-ch
}
