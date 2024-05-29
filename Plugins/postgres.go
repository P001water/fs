package Plugins

import (
	"database/sql"
	"fmt"
	"fs/config"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

func PostgresScan(info *config.HostInfo) (tmperr error) {
	if config.NoBrute {
		return
	}
	starttime := time.Now().Unix()
	for _, user := range config.Userdict["postgresql"] {
		for _, pass := range config.Passwords {
			pass = strings.Replace(pass, "{user}", string(user), -1)
			flag, err := PostgresConn(info, user, pass)
			if flag == true && err == nil {
				return err
			} else {
				errlog := fmt.Sprintf("[-] psql %v:%v %v %v %v", info.Host, info.Ports, user, pass, err)
				config.LogError(errlog)
				tmperr = err
				if config.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(config.Userdict["postgresql"])*len(config.Passwords)) * config.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func PostgresConn(info *config.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := info.Host, info.Ports, user, pass
	dataSourceName := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", Username, Password, Host, Port, "postgres", "disable")
	db, err := sql.Open("postgres", dataSourceName)
	if err == nil {
		db.SetConnMaxLifetime(time.Duration(config.Timeout) * time.Second)
		defer db.Close()
		err = db.Ping()
		if err == nil {
			result := fmt.Sprintf("[+] Postgres:%v:%v:%v %v", Host, Port, Username, Password)
			config.LogSuccess(result)
			flag = true
		}
	}
	return flag, err
}
