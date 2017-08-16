package uptimer

import (
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type AppUptimer struct{}

const (
	PROC_UPTIME = "/proc/uptime"
)

var main_uptimer *AppUptimer

var readFile = ioutil.ReadFile // for Mock() in the test

func (u *AppUptimer) getUptimeString() string {
	buf, err := readFile(PROC_UPTIME)
	if err != nil {
		glog.Error("Can't read '%s'", PROC_UPTIME)
		return "0"
	}
	rv := strings.Split(string(buf), " ")[0]
	return rv
}

func (u *AppUptimer) GetFloat() float64 {
	sbuf := u.getUptimeString()
	rv, err := strconv.ParseFloat(sbuf, 64)
	if err != nil {
		glog.Error("Can't convert '%s' to float64", sbuf)
		return 0
	}
	return float64(rv)
}

func (u *AppUptimer) Get() uint64 {
	sbuf := u.getUptimeString()
	sbuf = strings.Split(sbuf, ".")[0]
	rv, err := strconv.ParseInt(sbuf, 10, 64)
	if err != nil {
		glog.Error("Can't convert '%s' to uint64", sbuf)
		return 0
	}
	return uint64(rv)
}

type Uptimer interface {
	getUptimeString() string
	GetFloat() float64
	Get() uint64
}

func NewUptimer() *AppUptimer {
	// Uptimer is a singletone
	return main_uptimer
}

func init() {
	if _, err := os.Stat(PROC_UPTIME); os.IsNotExist(err) || os.IsPermission(err) {
		glog.Fatalf("File '%s' does not exists or You have no permissons to read it.", PROC_UPTIME)
	}
	main_uptimer = new(AppUptimer)
}
