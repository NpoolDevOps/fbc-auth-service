package main

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	loginmysql "github.com/NpoolDevOps/fbc-devops-service/mysql"
	loginredis "github.com/NpoolDevOps/fbc-devops-service/redis"
	"net/http"
)

type LoginConfig struct {
}

type LoginServer struct {
}

func NewLoginServer(config LoginConfig) *LoginServer {
	server := &LoginServer{}
	return server
}

func (s *LoginServer) Run() {
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

	httpdaemon.Run(MyConfig.Port)
}
