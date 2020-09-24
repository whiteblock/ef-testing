package main

import (
	"strconv"
	"time"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"regexp"
)

var headExtr = regexp.MustCompile(`(?m)^[A-Z| ]{3,5}\[[0-9]*-[0-9]*\|[0-9]*\:[0-9]*:[0-9]*\.[0-9]*\]`)

type SyslogngOutput struct {
    Testrun         string  `json:"TESTRUN"`
    Test            string  `json:"TEST"`
    Priority        string  `json:"PRIORITY"`
    Phase           string  `json:"PHASE"`
    Org             string  `json:"ORG"`
    Name            string  `json:"NAME"`
    Message         string  `json:"MESSAGE"`
    ImageName       string  `json:"IMAGE_NAME"`
    ContainerTag    string  `json:"CONTAINER_TAG"`
    ContainerName   string  `json:"CONTAINER_NAME"`
}

// Geth logline
type LogLine struct {
    UnixNanoTime    int64                   `json:"unixNanoTime"`
    Level           string                  `json:"level"`
    Message         string                  `json:"message"`
    Values          map[string]interface{}  `json:"values`
}

type Outputter struct {
    Split        bool
    Destination  string
    outputs      map[string]io.WriteCloser
    singleOutput io.WriteCloser
    id           string
}

func (o *Outputter) Setup() error {
    var err error
    if o.Split {
        os.MkdirAll(o.Destination, 0777)
        o.outputs = map[string]io.WriteCloser{}
    } else {
        o.singleOutput, err = os.Create(o.Destination)
    }
    return err
}


func (o *Outputter) parseStart(name string, message string) error {
    r, _ := regexp.Compile("[0-9]+")
    timestamp := r.FindString(message)
    f, err := os.Create(filepath.Join(o.Destination, name))
    f.Write([]byte(timestamp))
    f.Write([]byte("\n"))
    return err
}

func (o *Outputter) extractKVPairs(msg SyslogngOutput) LogLine {
    line := msg.Message
    if len(line) < 67 {
        // Some log lines don't have k-v pairs at the end
        return LogLine{}
    }
    head := headExtr.FindString(line)
    message := strings.TrimSpace(line[len(head):67])
    tail := line[67:]
    tail = strings.TrimSpace(tail)
    kvPairs := strings.Split(tail," ")
    kv := map[string]interface{}{}
    for _, kvPair := range kvPairs {
        if !strings.Contains(kvPair,"=") {
            continue
        }
        parts := strings.SplitN(kvPair,"=",2)
        if strings.HasPrefix(parts[1],"0x") {
            kv[parts[0]] = parts[1]
        }else if strings.HasPrefix(parts[1],"\"") {
            kv[parts[0]] = strings.Trim(parts[1],"\"")
        }else{
            val, err := strconv.ParseFloat(parts[1],64)
            if err != nil {
                kv[parts[0]] = parts[1]
            }else{
                kv[parts[0]] = val
            }
        }
        
    }
    lvl := strings.Split(head,"[")[0]
    rawTS := head[len(lvl):]

    years := time.Now().Year()
    t, err := time.Parse("[01-02|15:04:05.000]",rawTS)
    t = t.AddDate(years, 0, 0)
    if err != nil {
        panic(err)
    }
    data := LogLine {
        UnixNanoTime: t.UnixNano(),
        Level: lvl,
        Message: message,
        Values: kv,
    }

    return data
}

func isAnnounceBlock(logline LogLine) bool {
    if logline.Message == "Announced block" {
        return true
    }
    return false
}

func isImportSegment(logline LogLine) bool {
    if logline.Message == "Imported new chain segment" {
        return true
    }
    return false
}

func isSealBlock(logline LogLine) bool {
    if logline.Message == "Successfully sealed new block" {
        return true
    }
    return false
}

func isReorg(logline LogLine) bool {
    if logline.Message == "Chain reorg detected" {
        return true
    }
    return false
}


func (o *Outputter) handleInput(msg SyslogngOutput, id string) error {
    if id != "" && msg.Testrun != id {
        // skip log lines from other tests IDs
        return nil
    }

    name := regexp.MustCompile(`geth-service[0-9]*$`)
    if !name.MatchString(msg.ContainerName) {
        return nil
    }
    logline := o.extractKVPairs(msg)

    if (isImportSegment(logline) || 
        isSealBlock(logline) ||
        isReorg(logline)) == false {
        // Not a useful log line
        return nil
    }

    return o.routeOutput(msg.ContainerName, logline)
}

// Write to file
func (o *Outputter) routeOutput(name string, logline LogLine) error {
    data, err := json.Marshal(logline)
    if err != nil {
        return err
    }
    if !o.Split {
        o.singleOutput.Write(data)
        o.singleOutput.Write([]byte("\n"))
        return nil
    }
    if _, exists := o.outputs[name]; !exists {
        o.outputs[name], err = os.Create(filepath.Join(o.Destination, name))
        if err != nil {
            return err
        }
    }

    // write only if valid json
    if data != nil {
        o.outputs[name].Write(data)
        o.outputs[name].Write([]byte("\n"))
    }
    return nil
}