package config

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func Parse(Info *HostInfo) {
	ParseInput(Info)
	ParseUser()
	ParsePass(Info)
	ParseScantype(Info)
}

func ParseUser() {
	var Usernames []string

	if Username == "" && Userfile == "" {
		return
	}
	if Username != "" {
		Usernames = strings.Split(Username, ",")
	}

	if Userfile != "" {
		users, err := ReadLinesFromFile(Userfile)
		if err == nil {
			for _, user := range users {
				if user != "" {
					Usernames = append(Usernames, user)
				}
			}
		}
	}
	Usernames = RemoveDuplicate(Usernames)

	for name := range Userdict {
		Userdict[name] = Usernames
	}
}

func ParsePass(Info *HostInfo) {
	var PwdList []string
	if Password != "" {
		passs := strings.Split(Password, ",")
		for _, pass := range passs {
			if pass != "" {
				PwdList = append(PwdList, pass)
			}
		}
		Passwords = PwdList
	}
	if Passfile != "" {
		passs, err := ReadLinesFromFile(Passfile)
		if err == nil {
			for _, pass := range passs {
				if pass != "" {
					PwdList = append(PwdList, pass)
				}
			}
			Passwords = PwdList
		}
	}
	if URL != "" {
		urls := strings.Split(URL, ",")
		TmpUrls := make(map[string]struct{})
		for _, url := range urls {
			if _, ok := TmpUrls[url]; !ok {
				TmpUrls[url] = struct{}{}
				if url != "" {
					Urls = append(Urls, url)
				}
			}
		}
	}
	if UrlFile != "" {
		urls, err := ReadLinesFromFile(UrlFile)
		if err == nil {
			TmpUrls := make(map[string]struct{})
			for _, url := range urls {
				if _, ok := TmpUrls[url]; !ok {
					TmpUrls[url] = struct{}{}
					if url != "" {
						Urls = append(Urls, url)
					}
				}
			}
		}
	}
	if PortFile != "" {
		ports, err := ReadLinesFromFile(PortFile)
		if err == nil {
			newport := ""
			for _, port := range ports {
				if port != "" {
					newport += port + ","
				}
			}
			ScanPorts = newport
		}
	}
}

func ReadLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("[-] failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("[-] error scanning file %s: %v", filename, err)
	}

	return lines, nil
}

func ParseInput(Info *HostInfo) {
	if Info.Host == "" && HostFile == "" && URL == "" && UrlFile == "" {
		os.Exit(0)
	}

	if BruteThread <= 0 {
		BruteThread = 1
	}

	if TmpSave == true {
		IsSave = false
	}

	if ScanPorts == DefaultPorts {
		ScanPorts += "," + webPort
	}

	if PortAdd != "" {
		if strings.HasSuffix(ScanPorts, ",") {
			ScanPorts += PortAdd
		} else {
			ScanPorts += "," + PortAdd
		}
	}

	if UserAdd != "" {
		user := strings.Split(UserAdd, ",")
		for a := range Userdict {
			Userdict[a] = append(Userdict[a], user...)
			Userdict[a] = RemoveDuplicate(Userdict[a])
		}
	}

	if PassAdd != "" {
		pass := strings.Split(PassAdd, ",")
		Passwords = append(Passwords, pass...)
		Passwords = RemoveDuplicate(Passwords)
	}

	if Socks5Proxy != "" && !strings.HasPrefix(Socks5Proxy, "socks5://") {
		if !strings.Contains(Socks5Proxy, ":") {
			Socks5Proxy = "socks5://127.0.0.1" + Socks5Proxy
		} else {
			Socks5Proxy = "socks5://" + Socks5Proxy
		}
	}
	if Socks5Proxy != "" {
		fmt.Println("Socks5Proxy:", Socks5Proxy)
		_, err := url.Parse(Socks5Proxy)
		if err != nil {
			fmt.Println("Socks5Proxy parse error:", err)
			os.Exit(0)
		}
		NoPing = true
	}
	if Proxy != "" {
		if Proxy == "1" {
			Proxy = "http://127.0.0.1:8080"
		} else if Proxy == "2" {
			Proxy = "socks5://127.0.0.1:1080"
		} else if !strings.Contains(Proxy, "://") {
			Proxy = "http://127.0.0.1:" + Proxy
		}
		fmt.Println("Proxy:", Proxy)
		if !strings.HasPrefix(Proxy, "socks") && !strings.HasPrefix(Proxy, "http") {
			fmt.Println("no support this proxy")
			os.Exit(0)
		}
		_, err := url.Parse(Proxy)
		if err != nil {
			fmt.Println("Proxy parse error:", err)
			os.Exit(0)
		}
	}

	if Hash != "" && len(Hash) != 32 {
		fmt.Println("[-] Hash is error,len(hash) must be 32")
		os.Exit(0)
	} else {
		var err error
		HashBytes, err = hex.DecodeString(Hash)
		if err != nil {
			fmt.Println("[-] Hash is error,hex decode error")
			os.Exit(0)
		}
	}
}

func ParseScantype(Info *HostInfo) {
	_, ok := PORTList[Scantype]
	if !ok {
		showmode()
	}
	if Scantype != "all" && ScanPorts == DefaultPorts+","+webPort {
		switch Scantype {
		case "wmiexec":
			ScanPorts = "135"
		case "wmiinfo":
			ScanPorts = "135"
		case "smbinfo":
			ScanPorts = "445"
		case "hostname":
			ScanPorts = "135,137,139,445"
		case "smb2":
			ScanPorts = "445"
		case "web":
			ScanPorts = webPort
		case "webonly":
			ScanPorts = webPort
		case "ms17010":
			ScanPorts = "445"
		case "cve20200796":
			ScanPorts = "445"
		case "portscan":
			ScanPorts = DefaultPorts + "," + webPort
		case "main":
			ScanPorts = DefaultPorts
		default:
			port, _ := PORTList[Scantype]
			ScanPorts = strconv.Itoa(port)
		}
		fmt.Println("-m ", Scantype, " start scan the port:", ScanPorts)
	}
}

func CheckErr(text string, err error, flag bool) {
	if err != nil {
		fmt.Println("Parse", text, "error: ", err.Error())
		if flag {
			if err != ParseIPErr {
				fmt.Println(ParseIPErr)
			}
			os.Exit(0)
		}
	}
}

func showmode() {
	fmt.Println("[-] The specified scan type does not exist")
	fmt.Println("-m")
	for name := range PORTList {
		fmt.Println("   [" + name + "]")
	}
	os.Exit(0)
}
