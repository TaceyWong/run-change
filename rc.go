package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

type RunChange struct {
	Opt        *Options
	Exclude    *regexp.Regexp
	ProcessEnv map[string]string
	Paths      []string
	LastRun    time.Time
}

// RunChange executes command when file is changed
func (rc *RunChange) RunCommand(theFile string) {
	if rc.Opt.RunOnce {
		// 如果文件存在且文件最后修改时间小于最后运行时间，则停止运行
		return
	}
	printMSG := ""
	if rc.Opt.VerboseMode > 0 {
		printMSG += ""
	}
	if rc.Opt.VerboseMode > 1 {
		printMSG += ""
	}
	if rc.Opt.VerboseMode > 2 {
		printMSG += ""
	}
	if printMSG != "" {
		fmt.Printf("==> %s <==\n", printMSG)
	}
	// 设置环境变量
	// 如果设置为静默模式，则将标准输出在设置为devnull
	// 执行命令
	// 更新最后运行时间

}

func (rc *RunChange) IsInterested(path string) bool {
	if rc.Exclude.Match([]byte(path)) {
		return false
	}
	for _, p := range rc.Paths {
		if path == p {
			return true
		}
	}
	// TODO: 获取目录名看是否在paths

	// TODO: 如归递归，一层层往上找
	return false
}

func (rc *RunChange) OnChange(event fsnotify.Event) {
	if rc.IsInterested(event.Name) {
		rc.RunCommand(event.Name)
	}
}

func (rc *RunChange) OnCreated(event fsnotify.Event) {
}

func (rc *RunChange) OnModified(event fsnotify.Event) {

}

func (rc *RunChange) OnMoved(event fsnotify.Event) {
}

func (rc *RunChange) OnChmod(event fsnotify.Event) {

}

func (rc *RunChange) OnDeleted(event fsnotify.Event) {

}

func (rc *RunChange) SetEnvVar() {

}

func (rc *RunChange) Run() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				switch {
				case event.Has(fsnotify.Create):
					rc.OnCreated(event)
				case event.Has(fsnotify.Write):
					rc.OnModified(event)
					rc.OnChange(event)
				case event.Has(fsnotify.Chmod):
					rc.OnChmod(event)
				case event.Has(fsnotify.Remove):
					rc.OnDeleted(event)
				case event.Has(fsnotify.Rename):
					rc.OnMoved(event)

				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add("/tmp")
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func Watch(opt *Options) {
	rc := RunChange{
		ProcessEnv: make(map[string]string),
	}
	rc.Opt = opt
	patterns := []string{
		//Vim swap files
		`\..*\.sw[px]*$`,
		// file creation test file 4913
		`4913$`,
		// backup files
		`.~$`,
		// git directories
		`\.git/?`,
		// __pycache__ directories
		`__pycache__/?`,
	}
	for index, pattern := range patterns {
		patterns[index] = "(.+/)?" + pattern
	}
	rc.Exclude = regexp.MustCompile(strings.Join(patterns, "|"))

	// 获取环境变量
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		rc.ProcessEnv[pair[0]] = pair[1]
	}

}

func SetupExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal,Bye!")
		os.Exit(0)
	}()
}

type Options struct {
	Recursive   bool
	VerboseMode int
	RunOnce     bool
	RunAtStart  bool
	QuietMode   bool
	Command     string
	Help        bool
	Files       []string
}

var Usage = func() *Options {
	opt := Options{}
	flag.BoolVar(&opt.Recursive, "r", false, "Watch recursively")
	flag.IntVar(&opt.VerboseMode, "v", 0, "Verbose output")
	flag.BoolVar(&opt.RunOnce, "1", false, "Don't re-run command if files changed while command was running")
	flag.BoolVar(&opt.RunAtStart, "s", false, "Run command immediately at start")
	flag.BoolVar(&opt.QuietMode, "q", false, "Run command quietly")
	flag.StringVar(&opt.Command, "c", "", "command to execute")
	flag.BoolVar(&opt.Help, "h", false, "Show this help info")
	flag.Parse()

	helpInfo := `Run-Change(%s) - run a command when a file is changed

Usage: %s [-v-r-s-1-c] FILE/DIR COMMAND...
       %s [-v-r-s-1-c] FILE/DIR [FILE/DIR ...] -c COMMAND

Options:

`
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, helpInfo, name, name, name)
	flag.PrintDefaults()
	fmt.Println(`
Environment variables:

  - WHEN_CHANGED_EVENT: reflects the current event type that occurs.
		Could be either: file_created, file_modified, file_moved, file_deleted
  - WHEN_CHANGED_FILE: provides the full path of the file that has generated the event.

Copyright (c) 2023, Tacey Wong.
License: MIT, see LICENSE for more details.`)
	return &opt
}

func main() {
	SetupExit()
	Watch(Usage())
}
