package main

import (
    "fmt"
    "os"
	"os/exec"
	_ "net"
	_"strings"
    "path/filepath"
	log "github.com/sirupsen/logrus"
)

func getExternalIP() (string, error){
	cmd := exec.Command("curl", "-s", "ifconfig.so")
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": err}).Error("Unable to set external IP.")
		return "", fmt.Errorf("Unable to set genesis external IP.")

	} else {
/*		fmt.Println(out)*/
		return string(out), nil
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

func getYamlFiles(file_path string) []string {

    var files []string

    err := filepath.Walk(file_path, func(path string, info os.FileInfo, err error) error {
        files = append(files, path)
        return nil
    })
    if err != nil {
        panic(err)
    }
    return files

}

func main() {
	ip, err := getExternalIP()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to get external IP exiting now.")
		return 
	}

	fmt.Println(ip)	
	files_yaml := getYamlFiles("/var/log/syslog-ng/ef-testing/test-yaml")
    for _, file := range files_yaml {
        fmt.Println(file)
    }

}