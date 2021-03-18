package main

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	authmysql "github.com/NpoolDevOps/fbc-devops-service/mysql"
	authredis "github.com/NpoolDevOps/fbc-devops-service/redis"
	fbclib "github.com/NpoolDevOps/fbc-license-service/library"
	"github.com/NpoolRD/http-daemon"
	"io/ioutil"
	"net/http"
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
	return nil, "", 0
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
