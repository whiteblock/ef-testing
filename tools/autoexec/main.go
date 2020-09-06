package main

import (
    "fmt"
    "os"
	"os/exec"
	"io/ioutil"
	"time"
	"syscall"
	_"bufio"
	"net/http"
	"encoding/json"
	_ "net"
	"strings"
    "path/filepath"
	log "github.com/sirupsen/logrus"
)

type TestEnv struct {
	HostName 	 string `json:"hostName"`
	TestName 	 string `json:"testName"`
	TimeBegin 	 int64	`json:"timeBegin"`
	TimeEnd 	 int64	`json:"timeEnd"`
	TestID		 string	`json:"testID"`
	WebStats	 string	`json:"webStats"`
}

type SysEnv struct {
	timeBegin 	 int64
	timeEnd 	 int64
	hostName 	 string
	fileName 	 string
	testID		 string
	UserName	 string
	webDataURL	 string
	pathSyslogNG string
	efLog  		 string
	pathLog  	 string
	autoExecLog	 string
	pathYaml	 string
	externalIP	 string
	webStats	 string
	rstatsPID	 int
}

func (test *SysEnv) setDefaults() error {
	test.hostName, _  = os.Hostname()
	test.fileName     = ""
	test.pathSyslogNG = "/var/log/syslog-ng/"
	test.efLog 		  = "ef-test.log"
	test.pathLog 	  = "autoexec-log/"
	test.autoExecLog  = "autoexec.log"
	test.pathYaml	  = "/var/log/syslog-ng/ef-testing/autoexec-yaml"
	test.externalIP	  = ""
/*	test.webStats	  = `{"difficulty":{"max":55000000,"standardDeviation":427516.00232242176,"mean":51202982.53106213},"totalDifficulty":{"max":25550288283,"standardDeviation":7391818514.634528,"mean":12768191452.43888},"gasLimit":{"max":11850000,"standardDeviation":595.7486135081332,"mean":11849961.018036073},"gasUsed":{"max":11371336,"standardDeviation":508257.5740591193,"mean":11342176.352705412},"blockTime":{"max":631,"standardDeviation":30.57527317048098,"mean":13.160965794768613},"blockSize":{"max":163758,"standardDeviation":7276.488035583049,"mean":162884.8316633266},"transactionPerBlock":{"max":25,"standardDeviation":1.1180317437084863,"mean":24.949899799599194},"uncleCount":{"max":1,"standardDeviation":0.09959738388608554,"mean":0.010020040080160317},"tps":{"max":25,"standardDeviation":7.716263384080033,"mean":6.520443852228358},"blocks":499}`*/
	test.webStats	  = ""
	
	nglog_file := test.pathSyslogNG+test.efLog
	if _, err := os.Stat(nglog_file); os.IsNotExist(err) {
    	f, _ := os.Create(nglog_file)
    	f.Close()
	}
	autoexeclog_file := test.pathSyslogNG+test.pathLog+test.autoExecLog
	if _, err := os.Stat(autoexeclog_file); os.IsNotExist(err) {
		err := os.Mkdir(test.pathSyslogNG+test.pathLog, 0755)
    	err = ioutil.WriteFile(autoexeclog_file, []byte(""), 0644)
    	if err != nil {
			log.WithFields(log.Fields{"file": autoexeclog_file, "error": err}).Error("Unable to create autoexeclog_file file .")
			return fmt.Errorf("Unable to create autoexeclog_file file.")
    	}
	}
	
return nil
}

func (test *SysEnv) getGenesisUserName() error {
	cmd := exec.Command("genesis", "whoami", "--json")
    out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": err}).Error("Unable to call genesis whoami.")
		return fmt.Errorf("Unable to call genesis whoami.")
	} 
	json.Unmarshal([]byte(out), &test)
	
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

func (test *SysEnv) startRstats() error {
	// genesis stats cad --json -t 670e1118-5620-4c91-8560-eafd14c73048  >> /var/log/syslog-ng/autoexec-log/670e1118-5620-4c91-8560-eafd14c73048.stats
	cmd := exec.Command("genesis", "stats", "cad", "-t", test.testID, "--json", ">>", test.pathSyslogNG+test.pathLog+test.testID+".stats")
	err := cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{"cmd": cmd, "error": err}).Error("Unable to start RSTATS process.")
		return fmt.Errorf("Unable to start RSTATS process.")
	} 
	fmt.Println(cmd)
	test.rstatsPID = cmd.Process.Pid
	
/*
	go func() { 
			err := cmd.Wait()
			log.WithFields(log.Fields{"cmd": cmd, "error": err}).Error("Unable to execute RSTATS process.")
		}()
*/
return nil
}

func (test *SysEnv) beginTest() error {
	cmd := exec.Command("genesis", "run", test.fileName, test.UserName, "--no-await", "--json")
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

func (test *SysEnv) cleanUp(test_err int) error {
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
	ok = syscall.Kill(test.rstatsPID, syscall.SIGKILL)
	if ok != nil {
		log.WithFields(log.Fields{"error": ok}).Error("Unable to kill RSTATS process.")
		return fmt.Errorf("Unable to kill RSTATS process TESTID: "+test.testID)

	} 
	// if err == 0 copy syslogng-logs to test-ef/ directory
	// cp ef-test.log test-ef/ef-test-670e1118.log
	if test_err == 0 {
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
			log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": ok}).Error("Unable to create new NG log file for current genesis test: "+test.testID)
			return fmt.Errorf("Unable to create new NG log file for current genesis test: "+test.testID)

		} 
	} else { // there mus have been an error so clear the NG log data

		// clear data from current syslog-ng/ef-test.log file
		cmd = exec.Command(">", test.pathSyslogNG+test.efLog)
		fmt.Println(cmd)
		out, ok = cmd.CombinedOutput()
		if ok != nil {
			log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": ok}).Error("Unable to clear NG logs for current genesis test: "+test.testID)
			return fmt.Errorf("Unable to clear NG logs for current genesis test: "+test.testID)

		} 
	}
	
	// Write test.webStats data to file
	file, ok := os.OpenFile(test.pathSyslogNG+test.pathLog+test.autoExecLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ok != nil {
		log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": ok}).Error("Unable to open autoexec log for current genesis test: "+test.testID)
		return fmt.Errorf("Unable to open autoexec log for current genesis test: "+test.testID)
	}
	split := strings.Split(test.fileName, "/")
	slice_file := strings.Split(split[len(split)-1], ".")
	
	time_now := time.Now()
	logStats := TestEnv{}
	logStats.HostName 	 = test.hostName
	logStats.TestName 	 = slice_file[0]
	logStats.TimeBegin 	 = test.timeBegin
	logStats.TimeEnd 	 = time_now.Unix()
	logStats.TestID		 = test.testID
	logStats.WebStats	 = test.webStats

	jsonTest, _ := json.Marshal(logStats)
	fmt.Println(string(jsonTest))
	line := string(jsonTest)+"\n"
	if _, err := file.WriteString(line); err != nil {
		log.WithFields(log.Fields{"out": string(out), "cmd": cmd, "error": ok}).Error("Unable to write to autoexec log for current genesis test: "+test.testID)
		return fmt.Errorf("Unable to write to autoexec log for current genesis test: "+test.testID)
	}

	file.Close()
return nil
}

func main() {
	err := error(nil)
	var test = SysEnv{}
	err = test.setDefaults()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to set/get default values. exiting now.")
		return 
	}
	test.externalIP, err = getExternalIP()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to get external IP exiting now.")
		return 
	}
	// set genesis username for this server
	test.getGenesisUserName()

	file := getYamlFiles(test.pathYaml)
    for i := 1; i < len(file); i++ {
    	// Set genesis settings set syslogng-host
    	err = setSyslogng(test.externalIP)
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "External IP": test.externalIP, "error": err}).Error("Unable to set syslogng-host genesis paramater.")
			return 
    	}
		time_now := time.Now()
		test.timeBegin = time_now.Unix()
		test.fileName = file[i]
    	// Begin genesis run yaml_file test
    	err = test.beginTest()
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
    	// Start RSTATS data collection
    	test.startRstats()
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to start RSTATS collection.")
			return 
    	}
/*
    	// Start Webdata collection
    	err = test.monitorWebData()
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to monitor web stats.")
			return 
    	}
*/
		fmt.Println(test)
    	// Test has finished cleanup and get ready for the next test
    	err = test.cleanUp(0)
    	if err != nil {
			log.WithFields(log.Fields{"yaml_file": file, "error": err}).Error("Unable to clean up after test.")
			return 
    	}
        fmt.Println(file[i])
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