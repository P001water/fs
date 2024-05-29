package main

import (
	"fmt"
	"fs/Plugins"
	"fs/config"
	"os"
	"time"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("[-] no input")
		os.Exit(0)
	}
	var Info config.HostInfo
	start := time.Now()
	config.Flag(&Info)
	config.Parse(&Info)
	Plugins.Scan(Info)
	fmt.Printf("[*] 扫描结束, 耗时: %s\n", time.Since(start))
}
