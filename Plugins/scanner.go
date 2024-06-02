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
	WaitCheckHosts, err := config.ParseIP(info.Host, config.HostFile, config.NoHosts)
	if err != nil {
		fmt.Println("[-] No_target_host", err)
		return
	}
	// 加载 poc 模块
	lib.Inithttp()

	var threadChan = make(chan struct{}, config.Threads)
	var wg = sync.WaitGroup{}
	web := strconv.Itoa(config.PORTList["web"])
	ms17010 := strconv.Itoa(config.PORTList["ms17010"])

	var aliveHosts []string
	var aliveAddr []string
	// 存活主机探测
	if len(WaitCheckHosts) > 0 {
		if config.NoPing == false || config.Scantype == "icmp" {
			aliveHosts = CheckHostLive(WaitCheckHosts, config.Ping)
		}
		if config.Scantype == "icmp" {
			config.LogWG.Wait()
			return
		}
	}

	// 存活主机端口扫描
	if len(aliveHosts) > 0 {
		if config.Scantype == "webonly" || config.Scantype == "webpoc" {
			aliveAddr = NoPortScan(aliveHosts, config.ScanPorts)
		} else if config.Scantype == "hostname" {
			config.ScanPorts = "139"
			aliveAddr = NoPortScan(aliveHosts, config.ScanPorts)
		} else {
			aliveAddr = PortScan(aliveHosts, config.ScanPorts, config.Timeout)
			fmt.Println("[*] alive ports len is:", len(aliveAddr))
			if config.Scantype == "portscan" {

				config.LogWG.Wait()
				return
			}
		}

		if len(config.HostPort) > 0 {
			aliveAddr = append(aliveAddr, config.HostPort...)
			aliveAddr = config.RemoveDuplicate(aliveAddr)
			config.HostPort = nil
			fmt.Println("[*] aliveAddr len is:", len(aliveAddr))
		}

		var servicePorts []string //servicePorts := []string{"21","22","135"."445","1433","3306","5432","6379","9200","11211","27017"...}
		for _, port := range config.PORTList {
			servicePorts = append(servicePorts, strconv.Itoa(port))
		}

		fmt.Println("[*] start vulscan")
		for _, targetIP := range aliveAddr {
			info.Host, info.Ports = strings.Split(targetIP, ":")[0], strings.Split(targetIP, ":")[1]
			if config.Scantype == "all" || config.Scantype == "main" {
				switch {
				case info.Ports == "135":
					AddScan(info.Ports, info, &threadChan, &wg) //findnet
					if config.IsWmi {
						AddScan("1000005", info, &threadChan, &wg) //wmiexec
					}
				case info.Ports == "445":
					AddScan(ms17010, info, &threadChan, &wg) //ms17010
					//AddScan(info.ScanPorts, info, threadChan, &wg)  //smb
					//AddScan("1000002", info, threadChan, &wg) //smbghost
				case info.Ports == "9000":
					AddScan(web, info, &threadChan, &wg)        //http
					AddScan(info.Ports, info, &threadChan, &wg) //fcgiscan
				case IsContain(servicePorts, info.Ports):
					AddScan(info.Ports, info, &threadChan, &wg) //plugins scan
				default:
					AddScan(web, info, &threadChan, &wg) //webtitle
				}
			} else {
				actionType := strconv.Itoa(config.PORTList[config.Scantype])
				AddScan(actionType, info, &threadChan, &wg)
			}
		}
	}

	for _, url := range config.Urls {
		info.Url = url
		AddScan(web, info, &threadChan, &wg)
	}
	wg.Wait()
	config.LogWG.Wait()
	fmt.Printf("[+] 已完成 %v/%v\n", config.End, config.Num)
	close(config.ResultsChan)
}

var Mutex = &sync.Mutex{}

func AddScan(actionType string, info config.HostInfo, ch *chan struct{}, wg *sync.WaitGroup) {
	*ch <- struct{}{}
	wg.Add(1)
	go func() {
		Mutex.Lock()
		config.Num += 1
		Mutex.Unlock()
		ConvertFunc(&actionType, &info)
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
	infoSlice := []reflect.Value{reflect.ValueOf(info)}
	f.Call(infoSlice)
}

func IsContain(items []string, item string) bool {
	itemMap := make(map[string]struct{})
	for _, eachItem := range items {
		itemMap[eachItem] = struct{}{}
	}
	_, contains := itemMap[item]
	return contains
}
