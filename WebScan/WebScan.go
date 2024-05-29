package WebScan

import (
	"embed"
	"fmt"
	"fs/WebScan/lib"
	"fs/config"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed pocs
var Pocs embed.FS
var once sync.Once
var AllPocs []*lib.Poc

func WebScan(info *config.HostInfo) {
	once.Do(initpoc)
	var pocinfo = config.Pocinfo
	buf := strings.Split(info.Url, "/")
	pocinfo.Target = strings.Join(buf[:3], "/")

	if pocinfo.PocName != "" {
		Execute(pocinfo)
	} else {
		for _, infostr := range info.Infostr {
			pocinfo.PocName = lib.CheckInfoPoc(infostr)
			Execute(pocinfo)
		}
	}
}

func getRandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return config.UASlice[rand.Intn(len(config.UASlice))]
}

func Execute(PocInfo config.PocInfo) {
	req, err := http.NewRequest("GET", PocInfo.Target, nil)
	if err != nil {
		errlog := fmt.Sprintf("[-] webpocinit %v %v", PocInfo.Target, err)
		config.LogError(errlog)
		return
	}

	req.Header.Set("User-agent", getRandomUserAgent())
	req.Header.Set("Accept", config.Accept)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	if config.Cookie != "" {
		req.Header.Set("Cookie", config.Cookie)
	}
	pocs := filterPoc(PocInfo.PocName)
	lib.CheckMultiPoc(req, pocs, config.PocNum)
}

func initpoc() {
	if config.PocPath == "" {
		entries, err := Pocs.ReadDir("pocs")
		if err != nil {
			fmt.Printf("[-] init poc error: %v", err)
			return
		}
		for _, one := range entries {
			path := one.Name()
			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				if poc, _ := lib.LoadPoc(path, Pocs); poc != nil {
					AllPocs = append(AllPocs, poc)
				}
			}
		}
	} else {
		fmt.Println("[+] load poc from " + config.PocPath)
		err := filepath.Walk(config.PocPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil || info == nil {
					return err
				}
				if !info.IsDir() {
					if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
						poc, _ := lib.LoadPocbyPath(path)
						if poc != nil {
							AllPocs = append(AllPocs, poc)
						}
					}
				}
				return nil
			})
		if err != nil {
			fmt.Printf("[-] init poc error: %v", err)
		}
	}
}

func filterPoc(pocname string) (pocs []*lib.Poc) {
	if pocname == "" {
		return AllPocs
	}
	for _, poc := range AllPocs {
		if strings.Contains(poc.Name, pocname) {
			pocs = append(pocs, poc)
		}
	}
	return
}
