package Plugins

import (
	"fmt"
	"fs/config"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
)

func SmbScan2(info *config.HostInfo) (tmperr error) {
	if config.NoBrute {
		return nil
	}
	hasprint := false
	starttime := time.Now().Unix()
	hash := config.HashBytes
	for _, user := range config.Userdict["smb"] {
	PASS:
		for _, pass := range config.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err, flag2 := Smb2Con(info, user, pass, hash, hasprint)
			if flag2 {
				hasprint = true
			}
			if flag == true {
				var result string
				if config.Domain != "" {
					result = fmt.Sprintf("[+] SMB2 %v:%v:%v\\%v ", info.Host, info.Ports, config.Domain, user)
				} else {
					result = fmt.Sprintf("[+] SMB2 %v:%v:%v ", info.Host, info.Ports, user)
				}
				if len(hash) > 0 {
					result += "hash: " + config.Hash
				} else {
					result += pass
				}
				config.LogSuccess(result)
				return err
			} else {
				var errlog string
				if len(config.Hash) > 0 {
					errlog = fmt.Sprintf("[-] smb2 %v:%v %v %v %v", info.Host, 445, user, config.Hash, err)
				} else {
					errlog = fmt.Sprintf("[-] smb2 %v:%v %v %v %v", info.Host, 445, user, pass, err)
				}
				errlog = strings.Replace(errlog, "\n", " ", -1)
				config.LogError(errlog)
				tmperr = err
				if config.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(config.Userdict["smb"])*len(config.Passwords)) * config.Timeout) {
					return err
				}
			}
			if len(config.Hash) > 0 {
				break PASS
			}
		}
	}
	return tmperr
}

func Smb2Con(info *config.HostInfo, user string, pass string, hash []byte, hasprint bool) (flag bool, err error, flag2 bool) {
	conn, err := net.DialTimeout("tcp", info.Host+":445", time.Duration(config.Timeout)*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()
	initiator := smb2.NTLMInitiator{
		User:   user,
		Domain: config.Domain,
	}
	if len(hash) > 0 {
		initiator.Hash = hash
	} else {
		initiator.Password = pass
	}
	d := &smb2.Dialer{
		Initiator: &initiator,
	}

	s, err := d.Dial(conn)
	if err != nil {
		return
	}
	defer s.Logoff()
	names, err := s.ListSharenames()
	if err != nil {
		return
	}
	if !hasprint {
		var result string
		if config.Domain != "" {
			result = fmt.Sprintf("[*] SMB2-shares %v:%v:%v\\%v ", info.Host, info.Ports, config.Domain, user)
		} else {
			result = fmt.Sprintf("[*] SMB2-shares %v:%v:%v ", info.Host, info.Ports, user)
		}
		if len(hash) > 0 {
			result += "hash: " + config.Hash
		} else {
			result += pass
		}
		result = fmt.Sprintf("%v shares: %v", result, names)
		config.LogSuccess(result)
		flag2 = true
	}
	fs, err := s.Mount("C$")
	if err != nil {
		return
	}
	defer fs.Umount()
	path := `Windows\win.ini`
	f, err := fs.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	defer f.Close()
	flag = true
	return
	//bs, err := ioutil.ReadAll(f)
	//if err != nil {
	//	return
	//}
	//fmt.Println(string(bs))
	//return

}

//if info.Path == ""{
//}
//path = info.Path
//f, err := fs.OpenFile(path, os.O_RDONLY, 0666)
//if err != nil {
//	return
//}
//flag = true
//_, err = f.Seek(0, io.SeekStart)
//if err != nil {
//	return
//}
//bs, err := ioutil.ReadAll(f)
//if err != nil {
//	return
//}
//fmt.Println(string(bs))
//return
//f, err := fs.Create(`Users\Public\Videos\hello.txt`)
//if err != nil {
//	return
//}
//flag = true
//
//_, err = f.Write([]byte("Hello world!"))
//if err != nil {
//	return
//}
//
//_, err = f.Seek(0, io.SeekStart)
//if err != nil {
//	return
//}
//bs, err := ioutil.ReadAll(f)
//if err != nil {
//	return
//}
//fmt.Println(string(bs))
//return
