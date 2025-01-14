package logger

import (
	"encoding/json"
	"fmt"
	"github.com/Nerinyan/Nerinyan-APIV2/bodyStruct"
	"github.com/Nerinyan/Nerinyan-APIV2/utils"
	"github.com/Nerinyan/Nerinyan-APIV2/webhook"
	"github.com/jasonlvhit/gocron"
	"github.com/labstack/echo/v4"
	"github.com/pterm/pterm"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var file *os.File
var Ch = make(chan struct{}) //UpdateLogFile
var LOG = make(chan interface{})

const (
	logPath        = "./log"
	maxLogFileSize = int64(1024 * 1024 * 1024)
)

func init() {
	go func() {
		for elem := range LOG {
			fmt.Println(elem)
		}
	}()
	go func() {
		setLogFile()
		checkLogFileLimit()
		_ = gocron.Every(1).Days().At("00:00:00").Do(setLogFile)
		_ = gocron.Every(1).Hours().At("00:00").Do(checkLogFileLimit)
		<-gocron.Start()

	}()
	pterm.Info.Println("logfile Scheduler Started.")

}

func checkLogFileLimit() {
	checkDir()

	files, err := ioutil.ReadDir(logPath)
	if err != nil {
		pterm.Error.Println(err)
		return
	}

	sort.Slice(files, func(i, j int) (tf bool) {
		fii, _ := strconv.Atoi(strings.Split(files[i].Name(), ".")[0])
		fij, _ := strconv.Atoi(strings.Split(files[j].Name(), ".")[0])
		return fii > fij
	})
	var fileSize int64
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		ri, err := regexp.Match("^([0-9][.]log)$", []byte(f.Name()))
		if err != nil || ri {
			continue
		}
		fileSize += f.Size()

		if maxLogFileSize < fileSize {
			err := os.Remove(fmt.Sprintf("%s/%s", logPath, f.Name()))
			if err != nil {
				pterm.Error.Println(err)

			} else {
				pterm.Info.Printf("logfile %s Deleted.", f.Name())
			}

		}
	}
}

func checkDir() {
	if _, e := os.Stat(logPath); os.IsNotExist(e) {
		err := os.MkdirAll(logPath, 666)
		if err != nil {
			pterm.Error.Println(err)
			panic(err)
		}
	}
}

func setLogFile() {

	if file != nil {
		file.Close()
	}
	checkDir()

	fileName := fmt.Sprintf("%s/%s.log", logPath, time.Now().Format("060102"))
	pterm.Info.Println("SET LOG FILE: ", fileName)
	fpLog, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_SYNC, 0777)
	if err != nil {
		pterm.Error.Println(err)
	}
	file = fpLog

	log.SetOutput(file)
	Ch <- struct{}{}

}

func Error(c echo.Context, v *bodyStruct.ErrorStruct) (vv *bodyStruct.ErrorStruct) {

	var fname string
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "?"
		line = 0
		fname = "?"
	} else {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			fname = "?"
		} else {
			fname = fn.Name()
		}
	}

	fmt.Printf("%s:%d %s.\n", file, line, fname)
	v.Path = c.Request().RequestURI
	v.RequestId = c.Response().Header().Get("X-Request-ID")
	json.Unmarshal(*utils.ToJsonString(c.QueryParams()), &v.Args.QueryParam)
	json.Unmarshal(*utils.ToJsonString(c.Cookies()), &v.Args.Cookie)
	json.Unmarshal(*utils.ToJsonString(c.Request().Header), &v.Args.Header)
	//fmt.Println(string(*utils.ToJsonString(c.QueryParams())))

	z := *v
	go func() {
		z.SourceFile = fmt.Sprintf("%s:%d", file, line)
		webhook.DiscordError(&z)

		b, _ := json.Marshal(&z)
		pterm.Error.Println(time.Now().Format("2006-01-02 15:04:05"), string(b))
	}()

	//TODO DB 에 저장
	return v

}

//func Info(v *bodyStruct.ErrorStruct) (vv *bodyStruct.ErrorStruct) {
//	z := *v
//	_, file, line, ok := runtime.Caller(1)
//	if !ok {
//		file = "???"
//		line = 0
//	}
//	go func() {
//
//		z.SourceFile = fmt.Sprintf("%s:%d", file, line)
//		webhook.DiscordError(&z)
//
//		b, _ := json.Marshal(v)
//		pterm.Info.Println(time.Now().Format("2006-01-02 15:04:05"), string(b))
//	}()
//
//	//TODO DB 에 저장
//	return v
//
//}
