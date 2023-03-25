package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
)

type RunChange struct {
	ctx        *cli.Context
	Exclude    *regexp.Regexp
	ProcessEnv map[string]string
	Paths      []string
	LastRun    time.Time
}

// RunChange executes command when file is changed
func (rc *RunChange) RunCommand(theFile string) {
	if rc.ctx.Bool("run-onece") {
		// 如果文件存在且文件最后修改时间小于最后运行时间，则停止运行
		return
	}
	printMSG := ""
	verbose := rc.ctx.Int("verbose")
	if verbose > 0 {
		printMSG += ""
	}
	if verbose > 1 {
		printMSG += ""
	}
	if verbose > 2 {
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

func (rc *RunChange) SetEnvVar(name, value string) {
	rc.ProcessEnv["WHEN_CHANGED_"+strings.ToUpper(name)] = value
}

func (rc *RunChange) GetEnvVar(name string) string {
	return rc.ProcessEnv["WHEN_CHANGED_"+strings.ToUpper(name)]
}

func (rc *RunChange) Run() error {
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
	return err
}

func NewRunChange(files []string, command string, ctx *cli.Context) *RunChange {
	rc := RunChange{
		ProcessEnv: make(map[string]string),
	}
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
	return &rc
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

func main() {
	app := &cli.App{
		Name:  "run-change",
		Usage: "run a command when a file is changed",
		UsageText: `
run-change [-v-r-s-1-c] FILE/DIR [FILE/DIR ...] [COMMAND]`,
		Authors: []*cli.Author{{Name: "Tacey Wong", Email: "xinyong.wang@qq.com"}},
		Description: `Environment variables:

		- WHEN_CHANGED_EVENT: reflects the current event type that occurs.
			  Could be either: file_created, file_modified, file_moved, file_deleted
		- WHEN_CHANGED_FILE: provides the full path of the file that has generated the event.`,
		Copyright:       "© 2023 Tacey Wong License under MIT, see LICENSE for more details",
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Value:   false,
				Usage:   "Watch recursively",
			},
			&cli.IntFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   0,
				Usage:   "`num` of verbose output",
			},
			&cli.BoolFlag{
				Name:    "run-once",
				Aliases: []string{"1"},
				Value:   false,
				Usage:   "Don't re-run command \n\tif files changed while command was running",
			},
			&cli.BoolFlag{
				Name:    "run-at-start",
				Aliases: []string{"s"},
				Value:   false,
				Usage:   "Run command immediately at start",
			},
			&cli.BoolFlag{
				Name:    "quiete",
				Aliases: []string{"q"},
				Value:   false,
				Usage:   "Run command quietly",
			},
			&cli.StringFlag{
				Name:    "command",
				Aliases: []string{"c"},
				Usage:   "Commands to run",
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() < 1 {
				return cli.ShowAppHelp(cCtx)
			}
			fmt.Println(cCtx.Args(), cCtx.NArg())
			var command string
			files := []string{}
			if c := cCtx.String("command"); c != "" {
				command = c
				files = cCtx.Args().Slice()
			} else if cCtx.NArg() >= 2 {
				command = strings.Join(cCtx.Args().Slice()[1:], " ")
				files = append(files, cCtx.Args().First())
			}
			println(command, files[0])
			if command == "" || len(files) == 0 {
				return cli.ShowAppHelp(cCtx)
			}

			if len(files) > 1 {
				if cCtx.Int("verbose") > 0 {
					l := fmt.Sprintf("'%s'", strings.Join(files, "', '"))
					fmt.Printf("When %s changes,run `%s` \n", l, command)
				}

			} else {
				if cCtx.Int("verbose") > 0 {
					fmt.Printf("When '%s' changes,run `%s` \n", files[0], command)
				}
			}
			SetupExit()
			rc := NewRunChange(files, command, cCtx)
			rc.Run()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
