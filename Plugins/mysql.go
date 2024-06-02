package Plugins

import (
	"database/sql"
	"fmt"
	"fs/config"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

func MysqlScan(info *config.HostInfo) (tmperr error) {
	if config.NoBrute {
		return
	}
	startTime := time.Now().Unix()
	for _, user := range config.Userdict["mysql"] {
		for _, pass := range config.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := MysqlConn(info, user, pass)
			if flag == true && err == nil {
				return err
			} else {
				errLog := fmt.Sprintf("[-] mysql %v:%v %v %v %v", info.Host, info.Ports, user, pass, err)
				config.LogError(errLog)
				tmperr = err
				if config.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-startTime > (int64(len(config.Userdict["mysql"])*len(config.Passwords)) * config.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func MysqlConn(info *config.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := info.Host, info.Ports, user, pass
	dataSourceName := fmt.Sprintf("%v:%v@tcp(%v:%v)/mysql?charset=utf8&timeout=%v", Username, Password, Host, Port, time.Duration(config.Timeout)*time.Second)
	db, err := sql.Open("mysql", dataSourceName)
	if err == nil {
		db.SetConnMaxLifetime(time.Duration(config.Timeout) * time.Second)
		db.SetConnMaxIdleTime(time.Duration(config.Timeout) * time.Second)
		db.SetMaxIdleConns(0)
		defer db.Close()
		err = db.Ping()
		if err == nil {
			result := fmt.Sprintf("[+] mysql %v:%v %v %v", Host, Port, Username, Password)
			config.LogSuccess(result)
			flag = true
		}
	}
	return flag, err
}
