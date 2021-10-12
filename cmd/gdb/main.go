package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/myceliums/gdb/model"
	"github.com/myceliums/gdb/templater"
)

type Config struct {
	Pkg    string
	Config []byte
	Output string
}

func NewConfig() *Config {
	x := &Config{}

	flag.StringVar(&x.Pkg, `pkg`, `model`, `specifies the package name`)
	flag.StringVar(&x.Output, `o`, `model.gen.go`, `specifies the output`)

	if !flag.Parsed() {
		flag.Parse()
	}

	if len(flag.Args()) < 1 {
		fmt.Println("to few arguments")
		os.Exit(1)
	}

	cfg := flag.Args()[0]
	btz, err := ioutil.ReadFile(cfg)
	errExit(err, `error reading config file`)

	x.Config = btz

	return x
}

func main() {
	cfg := NewConfig()

	mdl, err := model.New(cfg.Config)
	errExit(err, `error reading this config`)

	f, err := os.OpenFile(cfg.Output, os.O_WRONLY|os.O_CREATE, 0644)
	errExit(err, `error writing file`)

	errExit(templater.WriteTemplate(f, cfg.Pkg, *mdl), `error writing file`)

	errExit(f.Close(), `error writing file`)
}

func errExit(err error, msg ...string) {
	if err != nil {
		fmt.Println(strings.Join(msg, ` `))
		os.Exit(1)
	}
}
