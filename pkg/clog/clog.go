package clog

import (
	"log"
	"os"
	"path"
)

var (
	C *log.Logger
)

// Initialize the logger
func init() {
	// path where we save file. just "filename.log" if you wish to save in local dir, Open  /tmp/go-server.log to read in current config
	LOGFILE := path.Join(os.TempDir(), "seer.log")

	// create if not existent, otherwise append
	f, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	LstdFlags := log.Ldate | log.Ltime | log.Lshortfile
	C = log.New(f, "Seer : ", LstdFlags)
}

// Wrapper functions
func Print(v ...interface{}) {
	C.Print(v...)
}

func Println(v ...interface{}) {
	C.Println(v...)
}

func Printf(format string, v ...interface{}) {
	C.Printf(format, v...)
}

func Fatal(v ...interface{}) {
	C.Fatal(v...)
}

func Fatalln(v ...interface{}) {
	C.Fatalln(v...)
}

func Fatalf(format string, v ...interface{}) {
	C.Fatalf(format, v...)
}

func Panic(v ...interface{}) {
	C.Panic(v...)
}

func Panicln(v ...interface{}) {
	C.Panicln(v...)
}

func Panicf(format string, v ...interface{}) {
	C.Panicf(format, v...)
}
