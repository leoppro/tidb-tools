package main

import (
	"flag"
	"fmt"
	"github.com/pingcap/tidb-tools/pkg/file_uploader"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/ngaut/log"
)

var (
	mydumper     = flag.String("mydumper", "mydumper", "MyDumper executable file path")
	uploaderArgs []string
	mydumperArgs []string
)

func main() {
	parserArgs()
	driver, err := file_uploader.NewAWSS3FileUploaderDriver("", "", "", "", "", "")
	if err != nil {
		log.Fatalf("AWS S3 Driver create failure, err: %#v", err)
		os.Exit(-1)
	}
	uploader := file_uploader.NewFileUploader("", 0, 0, driver)
	execMydumper(mydumperArgs...)
	uploader.WaitAndClose()
}

func execMydumper(args ...string) {
	cmd := exec.Command(*mydumper, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Errorf("Exec Mydumper failure %#v", err)
		os.Exit(-1)
	}
	err = cmd.Wait()
	if err != nil {
		log.Errorf("Exec Mydumper failure %#v", err)
		os.Exit(-1)
	}
}

func parserArgs() {
	osArgs := os.Args[1:]
	uploaderFlagNames := make(map[string]interface{})

	flag.VisitAll(func(fg *flag.Flag) {
		uploaderFlagNames[fg.Name] = 0
	})
	uploaderFlagValue := false
	for _, arg := range osArgs {
		if uploaderFlagValue {
			uploaderArgs = append(uploaderArgs, arg)
			uploaderFlagValue = false
			continue
		}
		argTrimd := strings.TrimFunc(arg, func(r rune) bool {
			return unicode.IsSpace(r) || r == '-'
		})
		if _, exist := uploaderFlagNames[argTrimd]; exist {
			uploaderArgs = append(uploaderArgs, arg)
			uploaderFlagValue = true
			continue
		}
		mydumperArgs = append(mydumperArgs, arg)
	}
	defaultUsage := flag.Usage
	flag.Usage = func() {
		defaultUsage()
		fmt.Fprint(flag.CommandLine.Output(), "Usage of Mydumper:\n")
		execMydumper("--help")
	}
	// Ignore errors; CommandLine is set for ExitOnError.
	flag.CommandLine.Parse(uploaderArgs)
}
