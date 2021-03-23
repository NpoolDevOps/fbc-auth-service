package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	authmysql "github.com/NpoolDevOps/fbc-auth-service/mysql"
	authredis "github.com/NpoolDevOps/fbc-auth-service/redis"
	types "github.com/NpoolDevOps/fbc-auth-service/types"
	fbclib "github.com/NpoolDevOps/fbc-license-service/library"
	"github.com/NpoolRD/http-daemon"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"time"
)

type AuthConfig struct {
	RedisCfg authredis.RedisConfig `json:"redis"`
	MysqlCfg authmysql.MysqlConfig `json:"mysql"`
	Port     int                   `json:"port"`
}

type AuthServer struct {
	config      AuthConfig
	authText    string
	redisClient *authredis.RedisCli
	mysqlClient *authmysql.MysqlCli
}

func NewAuthServer(configFile string) *AuthServer {
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot read file %v: %v", configFile, err)
		return nil
	}

	config := AuthConfig{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse file %v: %v", configFile, err)
		return nil
	}

	log.Infof(log.Fields{}, "create redis cli: %v", config.RedisCfg)
	redisCli := authredis.NewRedisCli(config.RedisCfg)
	if redisCli == nil {
		log.Errorf(log.Fields{}, "cannot create redis client %v: %v", config.RedisCfg, err)
		return nil
	}

	log.Infof(log.Fields{}, "create mysql cli: %v", config.MysqlCfg)
	mysqlCli := authmysql.NewMysqlCli(config.MysqlCfg)
	if mysqlCli == nil {
		log.Errorf(log.Fields{}, "cannot create mysql client %v: %v", config.MysqlCfg, err)
		return nil
	}

	server := &AuthServer{
		config:      config,
		authText:    fbclib.FBCAuthText,
		redisClient: redisCli,
		mysqlClient: mysqlCli,
	}

	log.Infof(log.Fields{}, "successful to create auth server")

	return server
}

func (s *AuthServer) UserLoginRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}

	err = req.ParseForm()
	if err != nil {
		return nil, err.Error(), -2
	}

	hasTarget := false
	targetUrl := ""

	if url, ok := req.Form["target"]; ok {
		hasTarget = true
		targetUrl = url[0]
	}

	input := types.UserLoginInput{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		return nil, err.Error(), -3
	}

	appId, err := s.mysqlClient.QueryAppId(input.AppId)
	if err != nil {
		return nil, err.Error(), -4
	}

	user, err := s.mysqlClient.QueryUserWithPassword(input.Username, input.Password)
	if err != nil {
		return nil, err.Error(), -5
	}

	if user.Id != appId.UserId {
		return nil, "app id is not belong to user id", -5
	}

	output := types.UserLoginOutput{}
	type MyClaims struct {
		Username string
		UserId   uuid.UUID
		jwt.StandardClaims
	}

	userKey := fmt.Sprintf("%v:%v", user.Id, input.AppId)

	userInfo, err := s.redisClient.QueryUserInfo(userKey)
	if err != nil {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyClaims{
			Username: input.Username,
			UserId:   user.Id,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			},
		})
		tokenStr, err := token.SignedString([]byte("asdfjkjkdfjsalkjlfdaskjl"))
		if err != nil {
			return nil, err.Error(), -6
		}

		authCode := sha256.Sum256([]byte(tokenStr))
		authCodeStr := hex.EncodeToString(authCode[0:])

		s.redisClient.InsertKeyInfo("user", userKey, authredis.UserInfo{
			AuthCode: authCodeStr,
		}, 24*time.Hour)

		s.redisClient.InsertKeyInfo("authcode", hex.EncodeToString(authCode[0:]), user, 24*time.Hour)
		output.AuthCode = authCodeStr
	} else {
		output.AuthCode = userInfo.AuthCode
	}

	if hasTarget {
		output.TargetUrl = targetUrl
	}

	return output, "", 0
}

func (s *AuthServer) UserLogoutRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	return nil, "", 0
}

func (s *AuthServer) UserInfoRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}

	input := types.UserInfoInput{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		return nil, err.Error(), -2
	}

	if input.AuthCode == "" {
		return nil, "auth code is must", -3
	}

	info, err := s.redisClient.QueryAuthInfo(input.AuthCode)
	if err != nil {
		return nil, err.Error(), -4
	}

	user, err := s.mysqlClient.QueryAuthUser(info.Username)
	if err != nil {
		return nil, err.Error(), -5
	}

	userInfo := types.UserInfoOutput{
		Id:          user.Id,
		VisitorOnly: true,
		SuperUser:   false,
	}

	super, err := s.mysqlClient.QuerySuperUser(user.Id)
	if err == nil {
		userInfo.VisitorOnly = super.Visitor
		userInfo.SuperUser = true
	}

	return userInfo, "", 0
}

func (s *AuthServer) Run() error {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.UserLoginAPI,
		Handler:  s.UserLoginRequest,
		Method:   "POST",
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.UserLogoutAPI,
		Handler:  s.UserLogoutRequest,
		Method:   "POST",
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.UserInfoAPI,
		Handler:  s.UserInfoRequest,
		Method:   "POST",
	})

	httpdaemon.Run(s.config.Port)
	return nil
}
