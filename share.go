package main

import (
	"log"
	"os"
)

var std = log.New(os.Stdout, "", log.LstdFlags)

var verbosity int = 2

func Vf(level int, format string, v ...interface{}) {
	if level <= verbosity {
		std.Printf(format, v...)
	}
}
func V(level int, v ...interface{}) {
	if level <= verbosity {
		std.Print(v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= verbosity {
		std.Println(v...)
	}
}

