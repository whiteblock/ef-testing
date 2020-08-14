package main


import(
	"encoding/json"
	"time"
	"os"
	"bufio"
	"strings"
	"strconv"
	"regexp"
)


type LogLine struct {
	Timestamp time.Time `json:"timestamp"`
	Level string `json:"level"`
	Message string `json:"message"`
	Values map[string]interface{} `json:"values`
}


func main() {
	fd, err := os.Open(os.Args[1])
	if err != nil{
		panic(err)
	}
	out, err := os.Create(os.Args[2])
	if err != nil{
		panic(err)
	}
	var headExtr = regexp.MustCompile(`(?m)^[A-Z| ]{3,5}\[[0-9]*-[0-9]*\|[0-9]*\:[0-9]*:[0-9]*\.[0-9]*\]`)
	rdr := bufio.NewReader(fd)
	for {
		line, err := rdr.ReadString(byte('\n'))
		if err != nil {
			break
		}

		head := headExtr.FindString(line)
		message := strings.TrimSpace(line[len(head):67])
		tail := line[67:]
		tail = strings.TrimSpace(tail)
		kvPairs := strings.Split(tail," ")
		kv := map[string]interface{}{}
		for _,kvPair := range kvPairs {
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
		t, err := time.Parse("[01-02|15:04:05.000]",rawTS)
		if err != nil {
			panic(err)
		}
		data, _ := json.Marshal(LogLine{
			Timestamp: t,
			Level: lvl,
			Message: message,
			Values: kv,
		})
		out.Write(data)
		out.Write([]byte("\n"))
	}
}