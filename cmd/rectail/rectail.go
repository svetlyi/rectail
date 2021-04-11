package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/svetlyi/rectail"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

var (
	config         = rectail.NewDefaultCfg()
	startWith      strSlice
	regexpsToWatch strSlice
)

func init() {
	flag.Int64Var(&config.MaxOffset, "max_offset", config.MaxOffset, "max offset from the end of the file for the first printing")
	flag.Int64Var(&config.DelayMillisecond, "delay", config.DelayMillisecond, "delay between files scanning")
	flag.StringVar(&config.LogPrefix, "log_prefix", config.LogPrefix, "just a log file prefixm stored in OS temp folder")
	flag.Var(
		&startWith,
		"start_with",
		`
Start with directories (which directories to scan).

Example:
rectail -start_with /foo/bar -start_with foo
`,
	)
	flag.Var(
		&regexpsToWatch,
		"regexps_to_watch",
		`
Regular expressions to match files and directories. 
The order should be the same as start_with entities.
If there is no regular expression for the start_with[i] 
entity, any folders/files (entities) will be added to be 
watched later. 

To avoid troubles of replacing special symbols like . by
current folder, enclose your regular expressions in quoutes, for example ".*".

Example:
rectail -start_with /foo/bar -start_with foo -regexps_to_watch "[0-9]+\.log"
`,
	)
}

func main() {
	flag.Parse()

	f, err := ioutil.TempFile("", config.LogPrefix)
	if err != nil {
		log.Fatalf("could not create a temp file: %v", err)
	}

	log.Println("logs are in", f.Name())

	logger := log.New(f, "", log.Ldate|log.Lmicroseconds)
	logger.Println("starting rectail...")

	fileUpdates := make(chan rectail.FileUpdate)
	updates := make(chan string)

	if len(startWith) == 0 {
		log.Fatalln("You should specify entities to start with (--help for more info)")
	}

	rt, err := rectail.NewRecTail(
		startWith,
		regexpsToWatch,
		updates,
		config.DelayMillisecond,
		config.MaxOffset,
		logger,
	)
	if err != nil {
		log.Fatalf("could not initialize rectail: %v", err)
	}
	rectailCtx, rectailCtxCancel := context.WithCancel(context.Background())

	go waitForInterrupt(rectailCtxCancel)
	defer rectailCtxCancel()

	go processResults(fileUpdates, updates)
	if err = rt.Watch(rectailCtx, fileUpdates); err != nil {
		logger.Fatalln(err)
	}
}

func processResults(fileUpdates chan rectail.FileUpdate, updates chan string) {
	for {
		select {
		case fileUpdate := <-fileUpdates:
			fmt.Println("=>", fileUpdate.FullFilePath, "<=")
			for _, line := range fileUpdate.Lines {
				fmt.Println(line)
			}
		case update := <-updates:
			fmt.Println(update)
		}
	}
}

func waitForInterrupt(fn func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		fn()
	}
}
