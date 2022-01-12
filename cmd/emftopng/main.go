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
	"unsafe"

	"github.com/pzinovkin/emf"
)

const VERSION = "0.2.0"

var errlog = log.New(os.Stderr, "emf: ", 0)

var (
	flagVersion = flag.Bool("version", false, "")
)

var usage = `EMF images converter

Usage: emftopng [inputfile]
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

	file, err := emf.ReadFile(fdata)
	if err != nil {
		errlog.Fatal(err)
	}

	img := file.Draw()

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
