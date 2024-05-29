package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// 数组去重 使用泛型来去除切片中的重复元素 go > 1.18
func RemoveDuplicate[T comparable](old []T) []T {
	result := make([]T, 0, len(old))
	seen := make(map[T]struct{})

	for _, item := range old {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// mapPortToIPs 是一个映射函数，用于根据端口映射IP
func MapPortToIPs(ipPortList []string) map[int][]string {
	PortsToipMapping := make(map[int][]string)
	// 遍历列表，解析IP和端口，并将端口添加到对应IP的列表中
	for _, ipPort := range ipPortList {
		parts := strings.Split(ipPort, ":")
		ip := parts[0]
		port, _ := strconv.Atoi(parts[1])
		PortsToipMapping[port] = append(PortsToipMapping[port], ip)
	}
	return PortsToipMapping
}

func MapIPToPorts(ipPortList []string) {
	ipToPorts := make(map[string][]int)
	// 遍历列表，解析IP和端口，并将端口添加到对应IP的列表中
	for _, ipPort := range ipPortList {
		parts := strings.Split(ipPort, ":")
		ip := parts[0]
		port, _ := strconv.Atoi(parts[1])
		ipToPorts[ip] = append(ipToPorts[ip], port)
	}

	for ip, ports := range ipToPorts {
		sort.Ints(ports)
		result := fmt.Sprintf(" %s: %v", ip, ports)
		LogSuccess(result)
	}
}
