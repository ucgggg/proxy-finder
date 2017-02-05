package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mt    sync.Mutex
	hello = []byte("GET http://www.baidu.com/cache/global/img/gs.gif HTTP/1.1\r\nHost: www.baidu.com\r\nUser-Agent: Mozilla\r\nIf-Modified-Since: Thu, 19 Apr 2012 09:51:20 GMT\r\n\r\n")
	fs    *os.File
)

func (w *WriteRs) tsProxy(address string) {
	ok := false
	conn, _ := net.DialTimeout("tcp", address, time.Second*10)
	if conn != nil {
		conn.SetReadDeadline(time.Now().Add(time.Second * 40))
		_, err := conn.Write(hello)
		if err == nil {
			buf := make([]byte, 12)
			n, _ := conn.Read(buf)
			ok = n == len(buf) && (string(buf[len(buf)-3:]) == "304")
			log.Println(w.LastPc+`% BD`, address, ok, string(buf))
		}
		conn.Close()
	}
	if ok {
		mt.Lock()
		fs.Write([]byte(address + "\n"))
		fs.Sync()
		mt.Unlock()
	}
	return
}

type WriteRs struct {
	Buf    string
	LastPc string
}

func RegGet(pm string, s string, idx int) string {
	exp, _ := regexp.Compile(pm)
	v := exp.FindStringSubmatch(s)
	if v != nil {
		return v[idx]
	}
	return ""
}

func RegGet0(pm string, s string) string {
	return RegGet(pm, s, 0)
}

func RegGet1(pm string, s string) string {
	return RegGet(pm, s, 1)
}

func RegAllSub(pm string, s string) [][]string {
	exp, _ := regexp.Compile(pm)
	v := exp.FindAllStringSubmatch(s, -1)
	if v != nil {
		return v
	}
	return [][]string{}
}

func RegAllSub1(pm string, s string, n int) []string {
	arr := RegAllSub(pm, s)
	rs := make([]string, len(arr))
	for i, v := range arr {
		rs[i] = v[n]
	}
	return rs
}

func (w *WriteRs) Write(p []byte) (n int, err error) {
	mt.Lock()
	s := string(p)
	if w.Buf != "" {
		s = w.Buf + s
	}
	lastpc := RegGet1(`([0-9]+)\.[0-9]+% done`, s)
	if lastpc != "" && w.LastPc != lastpc {
		w.LastPc = lastpc
		log.Println(RegGet0(`rate:.*found=[0-9]+`, s))
	}
	idx := strings.LastIndexAny(s, "\r\n")
	if idx >= 0 {
		sr := s[:idx]
		ips := RegAllSub1(`open port ([0-9]+)/tcp on ([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`, sr, 2)
		ports := RegAllSub1(`open port ([0-9]+)/tcp on ([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`, sr, 1)
		for i := 0; i < len(ips); i++ {
			go w.tsProxy(ips[i] + ":" + ports[i])
		}
		if idx+1 < len(s) {
			w.Buf = s[idx+1:]
		} else {
			w.Buf = ""
		}
	} else {
		w.Buf = s
	}
	mt.Unlock()
	return len(p), nil
}

func masscan(rate int, ports string) {
	cmdline := `-iL zone_include.zone -p ` + ports + ` --excludefile zone_exclude.zone --rate ` + strconv.Itoa(rate) + ` --seed 0`
	log.Println(cmdline)
	cmd := exec.Command("./masscan/bin/masscan", strings.Split(cmdline, " ")...)
	cmd.Stderr = new(WriteRs)
	cmd.Stdout = cmd.Stderr
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	cmd.Wait()
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: proxy-finder speed ports")
	}
	f, err := os.OpenFile("pf.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("create pf.log failed")
		return
	}
	fs = f
	//rate
	rate, _ := strconv.Atoi(os.Args[1])
	ports := os.Args[2]
	masscan(rate, ports)
}
