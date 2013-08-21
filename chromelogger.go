package chromelogger

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
)

const (
	version = "0.1"
	DEFAULT_CALL_DEPTH = 3
)

var testJSON = `
eyJ2ZXJzaW9uIjoiNC4xLjAiLCJjb2x1bW5zIjpbImxvZyIsImJhY2t0cmFjZSIsInR5cGUiXSwicm93cyI6W1tbImhlbGxvIGNvbnNvbGUiXSwiXC9ob21lXC93d3dcL3NvbW1hbi5zb21jbG91ZC5jb21cL2luZGV4LnBocCA6IDQiLCIiXSxbW3siSFRUUFMiOiJvbiIsIkhUVFBfSE9TVCI6InRlc3Quc29tbWFuLnNvbWNsb3VkLmNvbSIsIkhUVFBfQ09OTkVDVElPTiI6ImtlZXAtYWxpdmUiLCJIVFRQX0NBQ0hFX0NPTlRST0wiOiJtYXgtYWdlPTAiLCJIVFRQX0FDQ0VQVCI6InRleHRcL2h0bWwsYXBwbGljYXRpb25cL3hodG1sK3htbCxhcHBsaWNhdGlvblwveG1sO3E9MC45LCpcLyo7cT0wLjgiLCJIVFRQX1VTRVJfQUdFTlQiOiJNb3ppbGxhXC81LjAgKFdpbmRvd3MgTlQgNi4xOyBXT1c2NCkgQXBwbGVXZWJLaXRcLzUzNy4zNiAoS0hUTUwsIGxpa2UgR2Vja28pIENocm9tZVwvMjguMC4xNTAwLjk1IFNhZmFyaVwvNTM3LjM2IiwiSFRUUF9BQ0NFUFRfRU5DT0RJTkciOiJnemlwLGRlZmxhdGUsc2RjaCIsIkhUVFBfQUNDRVBUX0xBTkdVQUdFIjoia28tS1Isa287cT0wLjgsZW4tVVM7cT0wLjYsZW47cT0wLjQiLCJIVFRQX0NPT0tJRSI6Il9fdXRtYT0xODIzMjM0OTMuMTE2ODE2MzA4MC4xMzcxMjIyNTE1LjEzNzEyMjI1MTUuMTM3MTIyMjUxNS4xOyBfX3V0bXo9MTgyMzIzNDkzLjEzNzEyMjI1MTUuMS4xLnV0bWNzcj0oZGlyZWN0KXx1dG1jY249KGRpcmVjdCl8dXRtY21kPShub25lKTsgUEhQU0VTU0lEPXZsZTEwN2Y5N2g0NGduM2M5cTE2NGhkaDkxIiwiUEFUSCI6Ilwvc2JpbjpcL3Vzclwvc2JpbjpcL2JpbjpcL3VzclwvYmluIiwiU0VSVkVSX1NJR05BVFVSRSI6IiIsIlNFUlZFUl9TT0ZUV0FSRSI6IkFwYWNoZSIsIlNFUlZFUl9OQU1FIjoidGVzdC5zb21tYW4uc29tY2xvdWQuY29tIiwiU0VSVkVSX0FERFIiOiIxMjEuMjU0LjE3Ny4xMDUiLCJTRVJWRVJfUE9SVCI6IjQ0MyIsIlJFTU9URV9BRERSIjoiMjE4LjUyLjI0Ni4xOTciLCJET0NVTUVOVF9ST09UIjoiXC9ob21lXC93d3dcL3NvbW1hbi5zb21jbG91ZC5jb20iLCJTRVJWRVJfQURNSU4iOiJjbG91ZEB3emQuY29tIiwiU0NSSVBUX0ZJTEVOQU1FIjoiXC9ob21lXC93d3dcL3NvbW1hbi5zb21jbG91ZC5jb21cL2luZGV4LnBocCIsIlJFTU9URV9QT1JUIjoiNjE4MzEiLCJHQVRFV0FZX0lOVEVSRkFDRSI6IkNHSVwvMS4xIiwiU0VSVkVSX1BST1RPQ09MIjoiSFRUUFwvMS4xIiwiUkVRVUVTVF9NRVRIT0QiOiJHRVQiLCJRVUVSWV9TVFJJTkciOiIiLCJSRVFVRVNUX1VSSSI6IlwvaW5kZXgucGhwIiwiU0NSSVBUX05BTUUiOiJcL2luZGV4LnBocCIsIlBIUF9TRUxGIjoiXC9pbmRleC5waHAiLCJSRVFVRVNUX1RJTUUiOjEzNzYzMTU3MDd9XSwiXC9ob21lXC93d3dcL3NvbW1hbi5zb21jbG91ZC5jb21cL2luZGV4LnBocCA6IDUiLCIiXSxbWyJzb21ldGhpbmcgd2VudCB3cm9uZyEiXSwiXC9ob21lXC93d3dcL3NvbW1hbi5zb21jbG91ZC5jb21cL2luZGV4LnBocCA6IDYiLCJ3YXJuIl1dLCJyZXF1ZXN0X3VyaSI6IlwvaW5kZXgucGhwIn0=
`

type Row []interface{}

type Data struct {
	mu      sync.Mutex `json:"-"` // ensures atomic writes; protects the following fields
	out     http.ResponseWriter `json:"-"`
	Version string `json:"version"`
	Columns []string `json:"columns"`
	Rows    []Row `json:"rows"`
	//Rows    []interface{} `json:"rows"`
}

var data = NewData(nil)

func NewData(out http.ResponseWriter) *Data {
	return &Data{
		out: out,
		Version: version,
		Columns: []string{"log", "backtrace", "type"},
	}
}

// TODO Data.out이 직전 request의 http.ResponseWriter로 set 되어 있을 경우
// 두 번째 request에서 SetOutput을 하지 않으면 header에 다른 값(os.Stderr 이라도)을 넣어줄 수가 없음.
// 이럴 경우 기본 log로라도 남기고 싶은데, 이 Data.out이 비어있는지 아닌지 판단하기가 어려움
func (d *Data) log(logType string, args ...interface{}) {
	if d.out == nil {
		log.Println(args)
		return
	}

	// TODO 1. logs 얻고(convert로)
	//for _, v := range args {
	// 	row := d.convert(v)
	// 	d.Rows = append(d.Rows, row)
	//}
	logs := []interface{}{"test", "hello"}


	_, file, line, ok := runtime.Caller(DEFAULT_CALL_DEPTH)
	if !ok {
		file = "???"
		line = 0
	}
	bt := file + " : " + strconv.Itoa(line)

	d.addRow(logs, bt, logType)

}

func (d *Data) addRow(logs []interface{}, bt, logType string) {
	row := make(Row, 3)
	row[0] = logs
	row[1] = bt
	row[2] = logType
	d.Rows = append(d.Rows, row)
	d.writeHeader()
}

// TODO 제대로 된 값으로 완성해야 함
func (d *Data) convert(arg interface{}) Row {
	// 3 elements: Log, Backtrace, LogType
	row := make(Row, 3)


	row[0] = []interface{}{"hello console"}
	row[1] = "source line:5 "
	row[2] = "logType"

	return row
}

func (d *Data) writeHeader() {
	b, err := json.Marshal(d)

	//log.Println(string(b)) // XXX 지울 것
	log.Println(base64.StdEncoding.EncodeToString(b))

	if err != nil {
		log.Print(err)
		return
	}

	//data.out.Header().Set("X-ChromeLogger-Data", testJSON)
	data.out.Header().Set("X-ChromeLogger-Data", base64.StdEncoding.EncodeToString(b))
}





func (d *Data) Log(args ...interface{}) {
	w, ok := args[0].(http.ResponseWriter)
	if ok {
		data.out = w
		d.log("", args[1:]...)
	} else {
		d.log("", args...)
	}
}

func (d *Data) Logf(format string, args ...interface{}) {
	data.log("", fmt.Sprintf(format, args...))
}

func (d *Data) Warn(args ...interface{}) {
	w, ok := args[0].(http.ResponseWriter)
	if ok {
		data.out = w
		d.log("warn", args[1:]...)
	} else {
		d.log("warn", args...)
	}
}



// SetOutput sets the output destination for the default logger.
func SetOutput(w http.ResponseWriter) {
	data.mu.Lock()
	defer data.mu.Unlock()
	data.out = w
}

func Log(args ...interface{}) {
	data.Log(args...)
}

func Logf(format string, args ...interface{}) {
	data.Logf(format, args...)
}

func Warn(args ...interface{}) {
	data.Warn(args...)
}
