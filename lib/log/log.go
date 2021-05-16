package log

import (
	"log"
	"os"
)

var (
	Std = log.New(os.Stdout, "", log.LstdFlags)

	Verbosity int = 2
)

func Vf(level int, format string, v ...interface{}) {
	if level <= Verbosity {
		Std.Printf(format, v...)
	}
}
func V(level int, v ...interface{}) {
	if level <= Verbosity {
		Std.Print(v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= Verbosity {
		Std.Println(v...)
	}
}

