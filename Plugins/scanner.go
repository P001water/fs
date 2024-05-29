package Plugins

import (
	"fmt"
	"fs/WebScan/lib"
	"fs/config"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

func Scan(info config.HostInfo) {
	fmt.Println("[*] start_Live_scan")
	AllHosts, err := config.ParseIP(info.Host, config.HostFile, config.NoHosts)
	if err != nil {
		fmt.Println("[-] No_target_host", err)
		return
	}
	// 加载 poc 模块
	lib.Inithttp()

	var ch = make(chan struct{}, config.Threads)
	var wg = sync.WaitGroup{}
	web := strconv.Itoa(config.PORTList["web"])
	ms17010 := strconv.Itoa(config.PORTList["ms17010"])

	var AliveHosts []string
	var AliveAddr []string
	// 存活主机探测
	if len(AllHosts) > 0 {
		if config.NoPing == false || config.Scantype == "icmp" {
			AliveHosts = CheckHostLive(AllHosts, config.Ping)
		}
		if config.Scantype == "icmp" {
			config.LogWG.Wait()
			return
		}
	}

	// 存活主机端口扫描
	if len(AliveHosts) > 0 {
		if config.Scantype == "webonly" || config.Scantype == "webpoc" {
			AliveAddr = NoPortScan(AliveHosts, config.ScanPorts)
		} else if config.Scantype == "hostname" {
			config.ScanPorts = "139"
			AliveAddr = NoPortScan(AliveHosts, config.ScanPorts)
		} else {
			AliveAddr = PortScan(AliveHosts, config.ScanPorts, config.Timeout)
			fmt.Println("[*] alive ports len is:", len(AliveAddr))
			if config.Scantype == "portscan" {

				config.LogWG.Wait()
				return
			}
		}

		if len(config.HostPort) > 0 {
			AliveAddr = append(AliveAddr, config.HostPort...)
			AliveAddr = config.RemoveDuplicate(AliveAddr)
			config.HostPort = nil
			fmt.Println("[*] AliveAddr len is:", len(AliveAddr))
		}

		var serviceports []string //serviceports := []string{"21","22","135"."445","1433","3306","5432","6379","9200","11211","27017"...}
		for _, port := range config.PORTList {
			serviceports = append(serviceports, strconv.Itoa(port))
		}

		fmt.Println("[+] start vulscan")
		for _, targetIP := range AliveAddr {
			info.Host, info.Ports = strings.Split(targetIP, ":")[0], strings.Split(targetIP, ":")[1]
			if config.Scantype == "all" || config.Scantype == "main" {
				switch {
				case info.Ports == "135":
					AddScan(info.Ports, info, &ch, &wg) //findnet
					if config.IsWmi {
						AddScan("1000005", info, &ch, &wg) //wmiexec
					}
				case info.Ports == "445":
					AddScan(ms17010, info, &ch, &wg) //ms17010
					//AddScan(info.ScanPorts, info, ch, &wg)  //smb
					//AddScan("1000002", info, ch, &wg) //smbghost
				case info.Ports == "9000":
					AddScan(web, info, &ch, &wg)        //http
					AddScan(info.Ports, info, &ch, &wg) //fcgiscan
				case IsContain(serviceports, info.Ports):
					AddScan(info.Ports, info, &ch, &wg) //plugins scan
				default:
					AddScan(web, info, &ch, &wg) //webtitle
				}
			} else {
				scantype := strconv.Itoa(config.PORTList[config.Scantype])
				AddScan(scantype, info, &ch, &wg)
			}
		}
	}

	for _, url := range config.Urls {
		info.Url = url
		AddScan(web, info, &ch, &wg)
	}
	wg.Wait()
	config.LogWG.Wait()
	close(config.ResultsChan)
	fmt.Printf("[+] 已完成 %v/%v\n", config.End, config.Num)
}

var Mutex = &sync.Mutex{}

func AddScan(scantype string, info config.HostInfo, ch *chan struct{}, wg *sync.WaitGroup) {
	*ch <- struct{}{}
	wg.Add(1)
	go func() {
		Mutex.Lock()
		config.Num += 1
		Mutex.Unlock()
		ConvertFunc(&scantype, &info)
		Mutex.Lock()
		config.End += 1
		Mutex.Unlock()
		wg.Done()
		<-*ch
	}()
}

func ConvertFunc(name *string, info *config.HostInfo) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("[-] %v:%v scan error: %v\n", info.Host, info.Ports, err)
		}
	}()
	f := reflect.ValueOf(PluginList[*name])
	in := []reflect.Value{reflect.ValueOf(info)}
	f.Call(in)
}

func IsContain(items []string, item string) bool {
	itemMap := make(map[string]struct{})
	for _, eachItem := range items {
		itemMap[eachItem] = struct{}{}
	}
	_, contains := itemMap[item]
	return contains
}