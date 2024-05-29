package Plugins

import (
	"errors"
	"fmt"
	"fs/config"
	"github.com/stacktitan/smb/smb"
	"strings"
	"time"
)

func SmbScan(info *config.HostInfo) (tmperr error) {
	if config.NoBrute {
		return nil
	}
	starttime := time.Now().Unix()
	for _, user := range config.Userdict["smb"] {
		for _, pass := range config.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := doWithTimeOut(info, user, pass)
			if flag == true && err == nil {
				var result string
				if config.Domain != "" {
					result = fmt.Sprintf("[+] SMB %v:%v:%v\\%v %v", info.Host, info.Ports, config.Domain, user, pass)
				} else {
					result = fmt.Sprintf("[+] SMB %v:%v:%v %v", info.Host, info.Ports, user, pass)
				}
				config.LogSuccess(result)
				return err
			} else {
				errlog := fmt.Sprintf("[-] smb %v:%v %v %v %v", info.Host, 445, user, pass, err)
				errlog = strings.Replace(errlog, "\n", "", -1)
				config.LogError(errlog)
				tmperr = err
				if config.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(config.Userdict["smb"])*len(config.Passwords)) * config.Timeout) {
					return err
				}
			}
		}
	}
	return tmperr
}

func SmblConn(info *config.HostInfo, user string, pass string, signal chan struct{}) (flag bool, err error) {
	flag = false
	Host, Username, Password := info.Host, user, pass
	options := smb.Options{
		Host:        Host,
		Port:        445,
		User:        Username,
		Password:    Password,
		Domain:      config.Domain,
		Workstation: "",
	}

	session, err := smb.NewSession(options, false)
	if err == nil {
		session.Close()
		if session.IsAuthenticated {
			flag = true
		}
	}
	signal <- struct{}{}
	return flag, err
}

func doWithTimeOut(info *config.HostInfo, user string, pass string) (flag bool, err error) {
	signal := make(chan struct{})
	go func() {
		flag, err = SmblConn(info, user, pass, signal)
	}()
	select {
	case <-signal:
		return flag, err
	case <-time.After(time.Duration(config.Timeout) * time.Second):
		return false, errors.New("time out")
	}
}
