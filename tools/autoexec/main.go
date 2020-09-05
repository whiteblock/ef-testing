package main

import (
    "fmt"
    "os"
	"os/exec"
	"io/ioutil"
	"bufio"
	"net/http"
	"encoding/json"
	_ "net"
	_"strings"
    "path/filepath"
	log "github.com/sirupsen/logrus"
)

type SystemEnv struct {
	pathSyslogNG string "/var/log/syslog-ng/"
	efLog  		 string "ef-test.log"
	pathLog  	 string "test-ef"
	pathYaml	 string "/home/billhamilton/test/ef-testing/test-yaml"
	externalIP	 string ""
}

func getExternalIP() (string, error){
	cmd := exec.Command("curl", "-s", "ifconfig.so")
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": err}).Error("Unable to set external IP.")
		return "", fmt.Errorf("Unable to set genesis external IP.")

	} else {
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
		fmt.Println(string(out))
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

func monitorWebData(endpoint string) error {
	resp, err := http.Get("http://"+endpoint+":8080/stats/all")
	if err != nil {
		log.WithFields(log.Fields{"endpoint": endpoint, "resp": resp,"error": err}).Error("Unable to make http request.")
		return fmt.Errorf("Unable to make http request.")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)

	fmt.Println(result["blocks"])
	return nil
}

func beginTest(yaml_file string) (string, error) {
	genesis_out, _ := os.Open("genesis_out")
	defer genesis_out.Close()
	reader := bufio.NewReader(genesis_out)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println(line)	
	}	

	return "", nil
}

func main() {
	err := error(nil)
	sysEnv := SystemEnv{}
	sysEnv.externalIP, err = getExternalIP()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to get external IP exiting now.")
		return 
	}
/*
	files_yaml := getYamlFiles("/var/log/syslog-ng/ef-testing/test-yaml")
*/
	monitor_url := "skinnervenom-0.biomes.whiteblock.io"
	files_yaml := getYamlFiles(sysEnv.pathYaml)
    for _, file := range files_yaml {
    	err = setSyslogng(sysEnv.externalIP)
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to set syslogng-host genesis paramater.")
			return 
    	}
    	_ = monitorWebData(monitor_url)
/*
    	_, _ = beginTest(file)
*/
        fmt.Println(file)
    }

}