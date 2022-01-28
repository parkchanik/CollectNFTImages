package logger

import (
	"fmt"
	"io"
	"log"

	"os"

	"runtime"
	"time"
)

var imagelogging *log.Logger
var infologging *log.Logger

func LoggerInit() {

	var path string

	if runtime.GOOS == "windows" {
		path = ".\\log\\"
		fmt.Println("runtime.GOOS =", runtime.GOOS)
	} else if runtime.GOOS == "linux" {
		path = "./log/"
		fmt.Println("runtime.GOOS =", runtime.GOOS)
	}
	t := time.Now()
	//logfile := path + "transactiondata" + t.Format("20060102_15") + ".log"
	filenameprefix := "image"
	logfile := fmt.Sprintf("%s%s_%s.log", path, filenameprefix, t.Format("20060102_15"))
	datafile, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	imagelogging = log.New(io.MultiWriter(datafile, os.Stdout), "", 0)

	infofilenameprefix := "info"
	infologfile := fmt.Sprintf("%s%s_%s.log", path, infofilenameprefix, t.Format("20060102_15"))
	infofile, err := os.OpenFile(infologfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	infologging = log.New(io.MultiWriter(infofile, os.Stdout), "", 0)

}

func ImageLog(data string) {
	imagelogging.Println(data)
}

func InfoLog(format string, v ...interface{}) {
	infologging.Printf(format, v...)
}
