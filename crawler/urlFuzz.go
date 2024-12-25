package crawler

import (
	"fmt"
	"github.com/vicpsec/URLFinder/cmd"
	"github.com/vicpsec/URLFinder/config"
	"github.com/vicpsec/URLFinder/mode"
	"github.com/vicpsec/URLFinder/result"
	"github.com/vicpsec/URLFinder/util"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Fuzz
func UrlFuzz() {
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(cmd.U, 1)
	if len(hosts) == 0 {
		host = cmd.U
	} else {
		host = hosts[0]
	}
	if cmd.D != "" {
		host = cmd.D
	}
	disposes, _ := util.UrlDispose(append(result.ResultUrl, mode.Link{Url: cmd.U, Status: "200", Size: "0"}), host, "")
	if cmd.Z == 2 || cmd.Z == 3 {
		fuzz2(disposes)
	} else if cmd.Z != 0 {
		fuzz1(disposes)
	}
	fmt.Println("\rFuzz OK      ")
}

// fuzz请求
func fuzzGet(u string) {
	defer func() {
		config.Wg.Done()
		<-config.Ch
		util.PrintFuzz()
	}()
	if cmd.M == 3 {
		for _, v := range config.Risks {
			if strings.Contains(u, v) {
				return
			}
		}
	}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	if cmd.C != "" {
		request.Header.Set("Cookie", cmd.C)
	}
	//增加header选项
	request.Header.Set("User-Agent", util.GetUserAgent())
	request.Header.Set("Accept", "*/*")
	//加载yaml配置
	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}
	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		return
	} else {
		defer response.Body.Close()
	}
	code := response.StatusCode
	if strings.Contains(cmd.S, strconv.Itoa(code)) || cmd.S == "all" {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}
		body := string(dataBytes)
		re := regexp.MustCompile("<Title>(.*?)</Title>")
		title := re.FindAllStringSubmatch(body, -1)
		config.Lock.Lock()
		defer config.Lock.Unlock()
		if len(title) != 0 {
			result.Fuzzs = append(result.Fuzzs, mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Title: title[0][1], Source: "Fuzz"})
		} else {
			result.Fuzzs = append(result.Fuzzs, mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Title: "", Source: "Fuzz"})
		}

	}

}
func fuzz1(disposes []mode.Link) {
	dispose404 := []string{}
	for _, v := range disposes {
		if v.Status == "404" {
			dispose404 = append(dispose404, v.Url)
		}
	}
	fuzz1s := []string{}
	host := ""
	if len(dispose404) != 0 {
		host = regexp.MustCompile("(http.{0,1}://.+?)/").FindAllStringSubmatch(dispose404[0]+"/", -1)[0][1]
	}

	for _, v := range dispose404 {
		submatch := regexp.MustCompile("http.{0,1}://.+?(/.*)").FindAllStringSubmatch(v, -1)
		if len(submatch) != 0 {
			v = submatch[0][1]
		} else {
			continue
		}
		v1 := v
		v2 := v
		reh2 := ""
		if !strings.HasSuffix(v, "/") {
			_submatch := regexp.MustCompile("/.+(/[^/]+)").FindAllStringSubmatch(v, -1)
			if len(_submatch) != 0 {
				reh2 = _submatch[0][1]
			} else {
				continue
			}
		}
		for {
			re1 := regexp.MustCompile("/.+?(/.+)").FindAllStringSubmatch(v1, -1)
			re2 := regexp.MustCompile("(/.+)/[^/]+").FindAllStringSubmatch(v2, -1)
			if len(re1) == 0 && len(re2) == 0 {
				break
			}
			if len(re1) > 0 {
				v1 = re1[0][1]
				fuzz1s = append(fuzz1s, host+v1)
			}
			if len(re2) > 0 {
				v2 = re2[0][1]
				fuzz1s = append(fuzz1s, host+v2+reh2)
			}
		}
	}
	fuzz1s = util.UniqueArr(fuzz1s)
	config.FuzzNum = len(fuzz1s)
	config.Progress = 1
	fmt.Printf("Start %d Fuzz...\n", config.FuzzNum)
	fmt.Printf("\r                                           ")
	for _, v := range fuzz1s {
		config.Wg.Add(1)
		config.Ch <- 1
		go fuzzGet(v)
	}
	config.Wg.Wait()
	result.Fuzzs = util.Del404(result.Fuzzs)
}

func fuzz2(disposes []mode.Link) {
	disposex := []string{}
	dispose404 := []string{}
	for _, v := range disposes {
		if v.Status == "404" {
			dispose404 = append(dispose404, v.Url)
		}
		//防止太多跑不完
		if len(dispose404) > 20 {
			if v.Status != "timeout" && v.Status != "404" {
				disposex = append(disposex, v.Url)
			}
		} else {
			if v.Status != "timeout" {
				disposex = append(disposex, v.Url)
			}
		}

	}
	dispose, _ := util.PathExtract(disposex)
	_, targets := util.PathExtract(dispose404)

	config.FuzzNum = len(dispose) * len(targets)
	config.Progress = 1
	fmt.Printf("Start %d Fuzz...\n", len(dispose)*len(targets))
	fmt.Printf("\r                                           ")
	for _, v := range dispose {
		for _, vv := range targets {
			config.Wg.Add(1)
			config.Ch <- 1
			go fuzzGet(v + vv)
		}
	}
	config.Wg.Wait()
	result.Fuzzs = util.Del404(result.Fuzzs)
}
