package main

import (
	"flag"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/pzinovkin/emftoimg/emf"
)

const VERSION = "0.1.0"

var errlog = log.New(os.Stderr, "emf: ", 0)

var (
	flagVersion = flag.Bool("version", false, "")
)

var usage = `EMF images converter

Usage: emftoimg [inputfile]
   	--version  print the version number

`

func main() {

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	flag.Parse()

	if *flagVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	fname := flag.Arg(0)

	var fdata []byte
	var err error

	if fname != "" {
		fdata, err = ioutil.ReadFile(fname)
		if err != nil {
			errlog.Fatal(err)
		}
	}

	if fdata == nil && !isatty(os.Stdin.Fd()) {
		fdata, _ = ioutil.ReadAll(os.Stdin)
	}

	if fdata == nil {
		flag.Usage()
	}

	t1 := time.Now()
	file, err := emf.ReadFile(fdata)
	if err != nil {
		errlog.Fatal(err)
	}
	e1 := time.Since(t1)

	t2 := time.Now()
	img := file.Draw()
	e2 := time.Since(t2)

	var f io.Writer

	if fname != "" {
		f, err = os.Create(strings.TrimSuffix(fname, ".emf") + ".png")
		if err != nil {
			errlog.Fatal(err)
		}
		defer f.(*os.File).Close()

	} else {
		f = os.Stdout
	}

	err = png.Encode(f, img)
	if err != nil {
		errlog.Fatal(err)
	}

	errlog.Printf("file %d bytes reading %.3f ms conversion %.3f ms\n",
		len(fdata),
		float64(e1.Nanoseconds())/1000000,
		float64(e2.Nanoseconds())/1000000)

}

func isatty(fd uintptr) bool {
	var termios syscall.Termios

	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		uintptr(TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0,
		0,
		0)
	return err == 0
}
