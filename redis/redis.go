package authredis

import (
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	"github.com/go-redis/redis"
	"golang.org/x/xerrors"
	"time"
)

type RedisConfig struct {
	Host string        `json:"host"`
	Ttl  time.Duration `json:"ttl"`
}

type RedisCli struct {
	config RedisConfig
	client *redis.Client
}

func NewRedisCli(config RedisConfig) *RedisCli {
	cli := &RedisCli{
		config: config,
	}

	var myConfig RedisConfig

	resp, err := etcdcli.Get(config.Host)
	if err == nil {
		err = json.Unmarshal(resp[0], &myConfig)
		if err == nil {
			cli = &RedisCli{
				config: myConfig,
			}
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr: cli.config.Host,
		DB:   0,
	})

	log.Infof(log.Fields{}, "redis ping -> %v", config.Host)
	pong, err := client.Ping().Result()
	if err != nil {
		log.Errorf(log.Fields{}, "new redis client error [%v]", err)
		return nil
	}

	if pong != "PONG" {
		log.Errorf(log.Fields{}, "redis connect failed!")
	} else {
		log.Infof(log.Fields{}, "redis connect success!")
	}

	cli.client = client

	return cli
}

var redisKeyPrefix = "fbc:userauth:server:"

type UserInfo struct {
	AuthCode string `json:"auth_code"`
}

func (cli *RedisCli) InsertKeyInfo(keyWord string, id string, info interface{}, ttl time.Duration) error {
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%v:%v:%v", redisKeyPrefix, keyWord, id)
	val := string(b)
	log.Infof(log.Fields{}, "redis %v -> %v", key, val)

	err = cli.client.Set(key, val, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func (cli *RedisCli) QueryUserInfo(userKey string) (*UserInfo, error) {
	val, err := cli.client.Get(fmt.Sprintf("%v:user:%v", redisKeyPrefix, userKey)).Result()
	if err != nil {
		return nil, err
	}
	info := &UserInfo{}
	err = json.Unmarshal([]byte(val), info)
	if err != nil {
		return nil, err
	}
	if info.AuthCode == "" {
		return nil, xerrors.Errorf("invalid auth code")
	}
	return info, nil
}
