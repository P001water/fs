package Plugins

import (
	"bytes"
	"fmt"
	"fs/config"
	"golang.org/x/net/icmp"
	"net"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	AliveHosts []string
	ExistHosts = make(map[string]struct{})
	liveWG     sync.WaitGroup
)

// sortIPs sorts an array of IP addresses in ascending order
func sortIPs(ips []string) {
	sort.Slice(ips, func(i, j int) bool {
		ip1 := net.ParseIP(ips[i])
		ip2 := net.ParseIP(ips[j])

		// If either IP address is invalid, use string comparison
		if ip1 == nil || ip2 == nil {
			return ips[i] < ips[j]
		}

		// Compare the IP addresses using byte comparison
		return bytes.Compare(ip1.To16(), ip2.To16()) < 0
	})
}

func IPsortHandleWorker(Ping bool, WaitCheckHosts []string, AliveHostsChan chan string) {
	for ip := range AliveHostsChan {
		AliveHosts = append(AliveHosts, ip)
		liveWG.Done()
	}
	liveWG.Wait()
	sortIPs(AliveHosts)
	tmpList := AliveHosts
	if config.Silent == false {
		if Ping == false {
			for _, ip := range tmpList {
				result := fmt.Sprintf(" {icmp} Target %-15s is alive", ip)
				config.LogSuccess(result)
			}
		} else {
			for _, ip := range tmpList {
				result := fmt.Sprintf(" {ping} Target %-15s is alive", ip)
				config.LogSuccess(result)
			}
		}
	}
	tips := fmt.Sprintf("[*] Alive_Hosts: %d", len(tmpList))
	config.LogSuccess(tips)
}

func CheckHostLive(WaitCheckHosts []string, Ping bool) []string {

	// chan 接收存活主机并处理
	aliveHostsChan := make(chan string, len(WaitCheckHosts))
	go IPsortHandleWorker(Ping, WaitCheckHosts, aliveHostsChan)

	if Ping == false {
		// 优先尝试监听本地icmp,批量探测
		conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			config.LogError(err)
			//尝试无监听icmp探测
			fmt.Println("trying RunIcmpWithoutLst")
			conn, err := net.DialTimeout("ip4:icmp", "127.0.0.1", 3*time.Second)
			if err != nil {
				config.LogError(err)
				//使用ping探测
				fmt.Println("[-] The current user permissions unable to send icmp packets")
				fmt.Println("[-] start ping")
				RunPing(WaitCheckHosts, aliveHostsChan)
			} else {
				RunIcmpWithoutLst(WaitCheckHosts, aliveHostsChan)
			}
			defer conn.Close()
		}
		defer conn.Close()
		RunIcmpWithLst(WaitCheckHosts, conn, aliveHostsChan)
	}

	if Ping == true {
		//使用ping探测
		RunPing(WaitCheckHosts, aliveHostsChan)
	}

	liveWG.Wait()
	close(aliveHostsChan)

	if len(WaitCheckHosts) > 1000 {
		arrTop, arrLen := ArrayCountValueTop(AliveHosts, config.LiveTop, true)
		for i := 0; i < len(arrTop); i++ {
			output := fmt.Sprintf("[*] LiveTop %-16s 段存活数量为: %d", arrTop[i]+".0.0/16", arrLen[i])
			config.LogSuccess(output)
		}
	}

	if len(WaitCheckHosts) > 256 {
		arrTop, arrLen := ArrayCountValueTop(AliveHosts, config.LiveTop, false)
		for i := 0; i < len(arrTop); i++ {
			output := fmt.Sprintf("[*] LiveTop %-16s 段存活数量为: %d", arrTop[i]+".0/24", arrLen[i])
			config.LogSuccess(output)
		}
	}

	return AliveHosts
}

func RunIcmpWithLst(WaitCheckHosts []string, conn *icmp.PacketConn, aliveHostsChan chan string) {
	endFlag := false
	go func() {
		for {
			if endFlag == true {
				return
			}
			msg := make([]byte, 100)
			_, sourceIP, _ := conn.ReadFrom(msg)
			if sourceIP != nil {
				liveWG.Add(1)
				aliveHostsChan <- sourceIP.String()
			}
		}
	}()

	for _, host := range WaitCheckHosts {
		dst, _ := net.ResolveIPAddr("ip", host)
		IcmpByte := makeICMPEchoRequest(host)
		conn.WriteTo(IcmpByte, dst)
	}
	//根据hosts数量修改icmp监听时间
	start := time.Now()
	for {
		if len(AliveHosts) == len(WaitCheckHosts) {
			break
		}
		since := time.Since(start)
		var wait time.Duration
		switch {
		case len(WaitCheckHosts) <= 256:
			wait = time.Second * 3
		default:
			wait = time.Second * 6
		}
		if since > wait {
			break
		}
	}
	endFlag = true
	conn.Close()
}

func RunIcmpWithoutLst(hostslist []string, chanHosts chan string) {
	num := 1000
	if len(hostslist) < num {
		num = len(hostslist)
	}
	var wg sync.WaitGroup
	limiter := make(chan struct{}, num)
	for _, host := range hostslist {
		wg.Add(1)
		limiter <- struct{}{}
		go func(host string) {
			if icmpalive(host) {
				liveWG.Add(1)
				chanHosts <- host
			}
			<-limiter
			wg.Done()
		}(host)
	}
	wg.Wait()
	close(limiter)
}

func icmpalive(host string) bool {
	startTime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, 6*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	if err := conn.SetDeadline(startTime.Add(6 * time.Second)); err != nil {
		return false
	}
	msg := makeICMPEchoRequest(host)
	if _, err := conn.Write(msg); err != nil {
		return false
	}

	receive := make([]byte, 60)
	if _, err := conn.Read(receive); err != nil {
		return false
	}

	return true
}

func RunPing(hostslist []string, AliveHostsChan chan string) {
	var wg sync.WaitGroup
	limiter := make(chan struct{}, 50)
	for _, host := range hostslist {
		wg.Add(1)
		limiter <- struct{}{}
		go func(host string) {
			if ExecPingCmd(host) {
				liveWG.Add(1)
				AliveHostsChan <- host
			}
			<-limiter
			wg.Done()
		}(host)
	}
	wg.Wait()
}

func ExecPingCmd(ip string) bool {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("cmd", "/c", "ping -n 1 -w 1 "+ip+" && echo true || echo false")
		//ping -c 1 -i 0.5 -t 4 -W 2 -w 5 "+ip+" >/dev/null && echo true || echo false"
	case "darwin":
		command = exec.Command("/bin/bash", "-c", "ping -c 1 -W 1 "+ip+" && echo true || echo false")
		//ping -c 1 -i 0.5 -t 4 -W 2 -w 5 "+ip+" >/dev/null && echo true || echo false"
	default: //linux
		command = exec.Command("/bin/bash", "-c", "ping -c 1 -w 1 "+ip+" && echo true || echo false")
		//ping -c 1 -i 0.5 -t 4 -W 2 -w 5 "+ip+" >/dev/null && echo true || echo false"
	}
	outInfo := bytes.Buffer{}
	command.Stdout = &outInfo
	err := command.Start()
	if err != nil {
		return false
	}
	if err = command.Wait(); err != nil {
		return false
	} else {
		if strings.Contains(outInfo.String(), "true") && strings.Count(outInfo.String(), ip) > 2 {
			return true
		} else {
			return false
		}
	}
}

const (
	icmpHeaderSize = 8
	msgLength      = 40
)

func makeICMPEchoRequest(host string) []byte {
	msg := make([]byte, msgLength)

	// Set ICMP Type (8) and Code (0)
	msg[0], msg[1] = 8, 0

	// Generate Identifier and Sequence
	id0, id1 := genIdentifier(host)
	seq0, seq1 := genSequence(1)

	// Copy Identifier and Sequence to Message
	copy(msg[4:6], []byte{id0, id1})
	copy(msg[6:8], []byte{seq0, seq1})

	// Calculate and Set Checksum
	check := checkSum(msg[:icmpHeaderSize])
	msg[2], msg[3] = byte(check>>8), byte(check)

	return msg
}

func checkSum(msg []byte) uint16 {
	sum := uint32(0)
	length := len(msg)

	// 计算两个字节一组的和
	for i := 0; i < length-1; i += 2 {
		sum += uint32(msg[i])<<8 + uint32(msg[i+1])
	}

	// 处理奇数长度的消息
	if length%2 != 0 {
		sum += uint32(msg[length-1]) << 8
	}

	// 将进位加到结果的低 16 位上
	sum = (sum >> 16) + (sum & 0xffff)
	// 将进位再加到结果的低 16 位上
	sum += sum >> 16
	// 取反
	answer := uint16(^sum)
	return answer
}

func genSequence(v int16) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genIdentifier(host string) (byte, byte) {
	return host[0], host[1]
}

func ArrayCountValueTop(arrInit []string, length int, flag bool) (arrTop []string, arrLen []int) {
	if len(arrInit) == 0 {
		return
	}
	arrMap1 := make(map[string]int)
	arrMap2 := make(map[string]int)
	for _, value := range arrInit {
		line := strings.Split(value, ".")
		if len(line) == 4 {
			if flag {
				value = fmt.Sprintf("%s.%s", line[0], line[1])
			} else {
				value = fmt.Sprintf("%s.%s.%s", line[0], line[1], line[2])
			}
		}
		if arrMap1[value] != 0 {
			arrMap1[value]++
		} else {
			arrMap1[value] = 1
		}
	}
	for k, v := range arrMap1 {
		arrMap2[k] = v
	}

	i := 0
	for range arrMap1 {
		var maxCountKey string
		var maxCountVal = 0
		for key, val := range arrMap2 {
			if val > maxCountVal {
				maxCountVal = val
				maxCountKey = key
			}
		}
		arrTop = append(arrTop, maxCountKey)
		arrLen = append(arrLen, maxCountVal)
		i++
		if i >= length {
			return
		}
		delete(arrMap2, maxCountKey)
	}
	return
}
