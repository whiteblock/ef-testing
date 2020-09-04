package main

import (
    "fmt"
    "os"
	"os/exec"
	"net"
	_"strings"
    "path/filepath"
	log "github.com/sirupsen/logrus"
)

func GetOutboundIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				os.Stdout.WriteString(ipnet.IP.String() + "\n")
			}
		}
	}
}

func setSyslogng(ip string) error {

	cmd := exec.Command("genesis", "settings", "set", "syslogng-host", ip)
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"ip_address": ip, "out": string(out), "cmd": cmd, "error": err}).Error("Unable to set genesis syslogng host.")
		return fmt.Errorf("Unable to set genesis syslogng host.")

	} else {
		return nil
	}

}

func main() {
	ip := [][]string{
		[]string{"logs","logs-1"},
		[]string{"35.222.228.109","35.238.243.210"},
	}
	
/*
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
*/

	fmt.Println("hostname:", ip[0][0])

    var files []string

    root := "/var/log/syslog-ng/ef-testing/test-yaml"
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        files = append(files, path)
        return nil
    })
    if err != nil {
        panic(err)
    }
    for _, file := range files {
        fmt.Println(file)
    }
}