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

	input := types.UserLoginInput{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		return nil, err.Error(), -2
	}

	user, err := s.mysqlClient.QueryUserWithPassword(input.Username, input.Password)
	if err != nil {
		return nil, err.Error(), -3
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
			return nil, err.Error(), -4
		}

		authCode := sha256.Sum256([]byte(tokenStr))
		authCodeStr := hex.EncodeToString(authCode[0:])

		s.redisClient.InsertKeyInfo("user", userKey, authredis.UserInfo{
			AuthCode: authCodeStr,
		}, 24*time.Hour)

		input.Password = ""
		s.redisClient.InsertKeyInfo("authcode", hex.EncodeToString(authCode[0:]), input, 24*time.Hour)
		output.AuthCode = authCodeStr
	} else {
		output.AuthCode = userInfo.AuthCode
	}

	return output, "", 0
}

func (s *AuthServer) UserLogoutRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	return nil, "", 0
}

func (s *AuthServer) Run() error {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/api/v0/user/login",
		Handler:  s.UserLoginRequest,
		Method:   "POST",
	})

	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: "/api/v0/user/logout",
		Handler:  s.UserLogoutRequest,
		Method:   "POST",
	})

	httpdaemon.Run(s.config.Port)
	return nil
}
