package toolkit

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

func ReadLines(filename string) ([]string, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0400)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}

var GetSelf = func() (string, error) {
	return os.Args[0], nil
}

var TryWX = func() (fd *os.File, path string, err error) {

	paths := make([]string, 0)
	paths = append(paths, os.TempDir())
	paths = append(paths, "/dev/shm")
	paths = append(paths, "/tmp")
	paths = append(paths, "/data/local/tmp")

	for _, v := range paths {
		trydir := v + string(os.PathSeparator) + fmt.Sprintf("%v", rand.Int63())
		fd, err = os.OpenFile(trydir, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
		if os.IsExist(err) {
			paths = append(paths, v)
			continue
		}
		if err == nil {
			// ok
			return fd, trydir, nil
		}
	}

	return nil, "", err
}

/*func Vf(level int, format string, v ...interface{}) {
	if level <= 6 {
		fmt.Printf(format, v...)
	}
}
func V(level int, v ...interface{}) {
	if level <= 6 {
		fmt.Print(v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= 6 {
		fmt.Println(v...)
	}
}*/
