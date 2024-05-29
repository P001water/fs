package Plugins

import (
	"fmt"
	"fs/config"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Addr struct {
	ip   string
	port int
}

func PortScan(hostslist []string, ports string, timeout int64) []string {
	var AliveAddress []string
	probePorts := config.ParseScanPort(ports)
	if len(probePorts) == 0 {
		fmt.Printf("[-] parse port %s error, please check your port format\n", ports)
		return AliveAddress
	}
	noPorts := config.ParseScanPort(config.NoPorts)
	if len(noPorts) > 0 {
		temp := map[int]struct{}{}
		for _, port := range probePorts {
			temp[port] = struct{}{}
		}

		for _, port := range noPorts {
			delete(temp, port)
		}

		var newDatas []int
		for port := range temp {
			newDatas = append(newDatas, port)
		}
		probePorts = newDatas
		sort.Ints(probePorts)
	}
	workers := config.Threads
	AddrsChan := make(chan Addr, 100)
	resultsChan := make(chan string, 100)
	var wg sync.WaitGroup

	//接收结果
	go func() {
		for found := range resultsChan {
			AliveAddress = append(AliveAddress, found)
			wg.Done()
		}
	}()

	// 消费者 - 多线程扫描
	for i := 0; i < workers; i++ {
		go func() {
			for addr := range AddrsChan {
				PortConnect(addr, resultsChan, timeout, &wg)
				wg.Done()
			}
		}()
	}

	// 生产者 - 添加扫描目标
	for _, host := range hostslist {
		for _, port := range probePorts {
			wg.Add(1)
			AddrsChan <- Addr{host, port}
		}
	}

	wg.Wait()
	config.MapIPToPorts(AliveAddress)
	close(AddrsChan)
	close(resultsChan)
	return AliveAddress
}

func PortConnect(addr Addr, respondingHosts chan<- string, adjustedTimeout int64, wg *sync.WaitGroup) {
	host, port := addr.ip, addr.port
	conn, err := config.WrapperTcpWithTimeout("tcp4", fmt.Sprintf("%s:%v", host, port), time.Duration(adjustedTimeout)*time.Second)
	if err == nil {
		defer conn.Close()
		address := host + ":" + strconv.Itoa(port)
		//result := fmt.Sprintf(" %s open", address)
		//config.LogSuccess(result)
		wg.Add(1)
		respondingHosts <- address
	}
}

func NoPortScan(hostslist []string, ports string) (AliveAddress []string) {
	probePorts := config.ParseScanPort(ports)
	noPorts := config.ParseScanPort(config.NoPorts)
	if len(noPorts) > 0 {
		temp := map[int]struct{}{}
		for _, port := range probePorts {
			temp[port] = struct{}{}
		}

		for _, port := range noPorts {
			delete(temp, port)
		}

		var newDatas []int
		for port, _ := range temp {
			newDatas = append(newDatas, port)
		}
		probePorts = newDatas
		sort.Ints(probePorts)
	}
	for _, port := range probePorts {
		for _, host := range hostslist {
			address := host + ":" + strconv.Itoa(port)
			AliveAddress = append(AliveAddress, address)
		}
	}
	return
}
