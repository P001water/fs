package Plugins

import (
	"fmt"
	"fs/config"
	"github.com/jlaffaye/ftp"
	"strings"
	"time"
)

func FtpScan(info *config.HostInfo) (tmperr error) {
	if config.NoBrute {
		return
	}
	startTime := time.Now().Unix()
	flag, err := FtpConn(info, "anonymous", "")
	if flag && err == nil {
		return err
	} else {
		errlog := fmt.Sprintf("[-] ftp %v:%v %v %v", info.Host, info.Ports, "anonymous", err)
		config.LogError(errlog)
		tmperr = err
		if config.CheckErrs(err) {
			return err
		}
	}

	for _, user := range config.Userdict["ftp"] {
		for _, pass := range config.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := FtpConn(info, user, pass)
			if flag && err == nil {
				return err
			} else {
				errLog := fmt.Sprintf("[-] ftp %v:%v %v %v %v", info.Host, info.Ports, user, pass, err)
				config.LogError(errLog)
				tmperr = err
				if config.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-startTime > (int64(len(config.Userdict["ftp"])*len(config.Passwords)) * config.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func FtpConn(info *config.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := info.Host, info.Ports, user, pass
	conn, err := ftp.DialTimeout(fmt.Sprintf("%v:%v", Host, Port), time.Duration(config.Timeout)*time.Second)
	if err == nil {
		err = conn.Login(Username, Password)
		if err == nil {
			flag = true
			result := fmt.Sprintf("[+] ftp %v:%v %v %v", Host, Port, Username, Password)
			dirs, err := conn.List("")
			//defer conn.Logout()
			if err == nil {
				if len(dirs) > 0 {
					for i := 0; i < len(dirs); i++ {
						if len(dirs[i].Name) > 50 {
							result += "\n   [->]" + dirs[i].Name[:50]
						} else {
							result += "\n   [->]" + dirs[i].Name
						}
						if i == 5 {
							break
						}
					}
				}
			}
			config.LogSuccess(result)
		}
	}
	return flag, err
}
