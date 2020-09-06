package main

import (
    "fmt"
    "os"
	"os/exec"
	"io/ioutil"
	"time"
	_"bufio"
	"net/http"
	"encoding/json"
	_ "net"
	"strings"
    "path/filepath"
	log "github.com/sirupsen/logrus"
)

type TestEnv struct {
	TimeBegin 	 string
	TimeEnd 	 string
	TestID		 string
	WebStats	 string	
}

type SysEnv struct {
	timeBegin 	 int64
	timeEnd 	 int64
	testID		 string
	webDataURL	 string
	pathSyslogNG string
	efLog  		 string
	pathLog  	 string
	pathYaml	 string
	externalIP	 string
	webStats	 string
	rstatsPID	 int
}

func (test *SysEnv) setDefaults() {
	test.pathSyslogNG = "/var/log/syslog-ng/"
	test.efLog 		  = "ef-test.log"
	test.pathLog 	  = "autoexec-log/"
	test.pathYaml	  = "/var/log/syslog-ng/ef-testing/autoexec-yaml"
	test.externalIP	  = ""
	test.webStats	  = ""
}

func (test *SysEnv) getGenesisAccount() error {

return nil
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
		return fmt.Errorf("Unable to set genesis syslogng host.", ip)

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

func (test *SysEnv) getTestDNS() error {
	cmd := exec.Command("genesis", "info", test.testID, "--json")
	fmt.Println(cmd)
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"test_ID": test.testID, "out": string(out), "cmd": cmd, "error": err}).Error("Unable to get test DNS.")
		return fmt.Errorf("Unable to get test DNS.")

	}

	var result map[string]interface{}
	json.Unmarshal([]byte(out), &result)
	instance := result["instances"].([]interface{})
	domain := instance[0].(map[string]interface{})
	test.webDataURL = domain["domain"].(string)+".biomes.whiteblock.io:8080/stats/all"
return nil
}

func (test *SysEnv) getTestId() error {
	cmd := exec.Command("genesis", "tests", "-l", "-1")
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"ip_address": test.externalIP, "out": string(out), "cmd": cmd, "error": err}).Error("Unable to set genesis syslogng host.")
		return fmt.Errorf("Unable to set genesis syslogng host.")

	}
	// strip the new line character from the test id
	test.testID = strings.TrimSuffix(string(out), "\n")
return nil
}

func (test *SysEnv) monitorWebData() error {
	statsURL := "http://"+test.webDataURL
	
	for {
		resp, err := http.Get(statsURL)
		if err != nil {
			log.WithFields(log.Fields{"endpoint": statsURL, "resp": resp,"error": err}).Error("Unable to make http request.")
			return fmt.Errorf("Unable to make http request.")
		}

		body, err := ioutil.ReadAll(resp.Body)
		if string(body) != "" {
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			if result["blocks"].(int) > 420 {
				// set test.webStats to write to file when test is done
				test.webStats = string(body)
				resp.Body.Close()
				break
			}

			resp.Body.Close()
		}
		time.Sleep(15 * time.Second)
	}
return nil
}

func (test *SysEnv) startRstats(done chan bool) error {
	// genesis stats cad --json -t 670e1118-5620-4c91-8560-eafd14c73048  >> /var/log/syslog-ng/test-ef/670e1118-5620-4c91-8560-eafd14c73048.stats
/*	cmd := exec.Command("genesis", "stats", "cad", "-t", test.testID, "--json", ">>", test.pathSyslogNG+test.pathLog+test.testID+".stats")*/
	cmd := exec.Command("ls", "-lah")
	fmt.Println(cmd)
    done <- true
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": err}).Error("Unable to set genesis syslogng host.")
		return fmt.Errorf("Unable to set genesis syslogng host.")

	} 
	test.rstatsPID = cmd.Process.Pid
return nil
}

func (test *SysEnv) beginTest(yaml_file string) error {
	cmd := exec.Command("genesis", "run", yaml_file, "paccode", "--no-await", "--json")
	fmt.Println(cmd)
/*
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"ip_address": ip, "out": string(out), "cmd": cmd, "error": err}).Error("Unable to set genesis syslogng host.")
		return fmt.Errorf("Unable to set genesis syslogng host.")

	} 
*/
return nil
}

func (test *SysEnv) cleanUp(err int) error {
	// stop current genesis test
	cmd := exec.Command("genesis", "stop", test.testID)
	fmt.Println(cmd)
    out, ok := cmd.CombinedOutput()
	if ok != nil {
		log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": ok}).Error("Unable to stop current genesis test.")
		return fmt.Errorf("Unable to stop current genesis test: "+test.testID)

	}
	
	// give the NG && RSTATS logs 30 seconds to catch up
	time.Sleep(30 * time.Second)
	
	// kill RSTATS collection
	// ***********************
	
	
	// if err == 0 copy syslogng-logs to test-ef/ directory
	// cp ef-test.log test-ef/ef-test-670e1118.log
	if err == 0 {
		splitID := strings.Split(test.testID, "-")
		logFile := test.pathSyslogNG+test.efLog
		statsFile := test.pathSyslogNG+test.pathLog+"ef-test-"+splitID[0]+".log"
		ok := os.Rename(logFile, statsFile)
		if ok != nil {
			log.WithFields(log.Fields{"logFile": logFile, "statsFile": statsFile, "error": ok}).Error("Unable to copy NGlog data to stats directory.")
			return fmt.Errorf("Unable to copy NGlog data to stats directory: "+test.testID)

		} 
		cmd = exec.Command("touch", test.pathSyslogNG+test.efLog)
		fmt.Println(cmd)
		out, ok = cmd.CombinedOutput()
		if ok != nil {
			log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": err}).Error("Unable to create new NG log file for current genesis test: "+test.testID)
			return fmt.Errorf("Unable to create new NG log file for current genesis test: "+test.testID)

		} 
	} else { // there mus have been an error so clear the NG log data

		// clear data from current syslog-ng/ef-test.log file
		cmd = exec.Command(">", test.pathSyslogNG+test.efLog)
		fmt.Println(cmd)
		out, ok = cmd.CombinedOutput()
		if ok != nil {
			log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": err}).Error("Unable to clear NG logs for current genesis test: "+test.testID)
			return fmt.Errorf("Unable to clear NG logs for current genesis test: "+test.testID)

		} 
	}
	
	// Write test.webStats data to file
	jsonTest, _ := json.Marshal(&test)
	fmt.Println(string(jsonTest))
/*	fmt.Println(test)*/
return nil
}

func main() {
	err := error(nil)
	var test = SysEnv{}
	test.setDefaults()
	test.externalIP, err = getExternalIP()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to get external IP exiting now.")
		return 
	}
	file := getYamlFiles(test.pathYaml)
/*    for _, file := range files_yaml {*/
    for i := 1; i < len(file); i++ {
		done := make(chan bool, 1)
    	// Set genesis settings set syslogng-host
    	err = setSyslogng(test.externalIP)
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "External IP": test.externalIP, "error": err}).Error("Unable to set syslogng-host genesis paramater.")
			return 
    	}
		time_now := time.Now()
		test.timeBegin = time_now.Unix()
    	// Begin genesis run yaml_file test
    	err = test.beginTest(file[i])
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to begin test.")
			return 
    	}
    	// Get test ID
    	err = test.getTestId()
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to get test id.")
			return 
    	}
    	// Get test DNS so we can monitor web stats
    	err = test.getTestDNS()
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to get test web data URL.")
			return 
    	}
    	go test.startRstats(done)
/*
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to start RSTATS collection.")
			return 
    	}
*/
/*
    	err = test.monitorWebData()
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to monitor web stats.")
			return 
    	}
*/
		
  		<-done  	
    	err = test.cleanUp(0)
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to clean up after test.")
			return 
    	}
        fmt.Println(file)
    }
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
/*	log.SetFormatter(&log.JSONFormatter{})*/

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	/*	log.SetOutput(os.Stdout)*/

	log.SetReportCaller(true)
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}