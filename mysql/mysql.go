package authmysql

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/EntropyPool/entropy-logger"
	etcdcli "github.com/NpoolDevOps/fbc-license-service/etcdcli"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/xerrors"
)

type MysqlConfig struct {
	Host   string `json:"host"`
	User   string `json:"user"`
	Passwd string `json:"passwd"`
	DbName string `json:"db"`
}

type MysqlCli struct {
	config MysqlConfig
	url    string
	db     *gorm.DB
}

func NewMysqlCli(config MysqlConfig) *MysqlCli {
	cli := &MysqlCli{
		config: config,
		url: fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
			config.User, config.Passwd, config.Host, config.DbName),
	}

	var myConfig MysqlConfig

	resp, err := etcdcli.Get(config.Host)
	if err == nil {
		err = json.Unmarshal(resp[0], &myConfig)
		if err == nil {
			myConfig.DbName = config.DbName
			cli = &MysqlCli{
				config: myConfig,
				url: fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local",
					myConfig.User, myConfig.Passwd, myConfig.Host, myConfig.DbName),
			}
		}
	}

	log.Infof(log.Fields{}, "open mysql db %v", cli.url)
	db, err := gorm.Open("mysql", cli.url)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot open %v: %v", cli.url, err)
		return nil
	}

	log.Infof(log.Fields{}, "successful to create mysql db %v", cli.url)
	db.SingularTable(true)
	cli.db = db

	return cli
}

func (cli *MysqlCli) Delete() {
	cli.db.Close()
}

type AuthUser struct {
	Id       uuid.UUID `gorm:"column:id"`
	Username string    `gorm:"column:username"`
	Passwd   string    `gorm:"column:passwd"`
	Salt     string    `gorm:"column:salt"`
}

func (cli *MysqlCli) saltedPassword(password string, salt string) string {
	mac := hmac.New(sha256.New, []byte(salt))
	mac.Write([]byte(password))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum[0:])[0:12]
}

func (cli *MysqlCli) QueryUserWithPassword(username string, passwd string) (*AuthUser, error) {
	user := AuthUser{}

	var count int

	cli.db.Where("username = ?", username).Find(&user).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("user is not registered")
	}

	inputPasswd := passwd
	if user.Salt != "0" {
		inputPasswd = cli.saltedPassword(passwd, user.Salt)
	}

	if user.Passwd != inputPasswd {
		return nil, xerrors.Errorf("password is mismatched")
	}

	return &user, nil
}

func (cli *MysqlCli) QueryAuthUser(username string) (*AuthUser, error) {
	user := AuthUser{}

	var count int

	cli.db.Where("username = ?", username).Find(&user).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("user is not registered")
	}

	user.Passwd = ""

	return &user, nil
}

type AppId struct {
	Id     uuid.UUID `gorm:"column:id"`
	UserId uuid.UUID `gorm:"column:user_id"`
}

func (cli *MysqlCli) QueryAppId(id uuid.UUID) (*AppId, error) {
	appId := AppId{}
	var count int

	cli.db.Where("id = ?", id).Find(&appId).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("app id is not registered")
	}

	return &appId, nil
}

type SuperUser struct {
	Id      uuid.UUID `gorm:"column:id"`
	Visitor bool      `gorm:"column:visitor"`
}

func (cli *MysqlCli) QuerySuperUser(id uuid.UUID) (*SuperUser, error) {
	superUser := SuperUser{}
	var count int

	cli.db.Where("id = ?", id).Find(&superUser).Count(&count)
	if count == 0 {
		return nil, xerrors.Errorf("user is not super user")
	}

	return &superUser, nil
}
