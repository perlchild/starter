package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloud66/starter/common"
	"github.com/cloud66/starter/packs"
	"github.com/kardianos/osext"
)

var (
	flagPath         string
	flagTemplatePath string
	flagOverwrite    bool
	flagEnvironment  string
)

func init() {
	flag.StringVar(&flagPath, "p", "", "project path")
	flag.StringVar(&flagTemplatePath, "templates", "", "where template files are located")
	flag.BoolVar(&flagOverwrite, "o", false, "overwrite existing files")
	flag.StringVar(&flagEnvironment, "e", "production", "set project environment")
}

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "help" {
		flag.PrintDefaults()
		return
	}

	flag.Parse()

	fmt.Println(common.MsgTitle, "Cloud 66 Starter - (c) 2015 Cloud 66", common.MsgReset)

	if flagPath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("%s Unable to detect current directory path due to %s", common.MsgError, err.Error())
		}
		flagPath = pwd
	}

	if flagTemplatePath == "" {
		execDir, err := osext.Executable()
		if err != nil {
			fmt.Printf("%s Unable to detect template folder due to %s", common.MsgError, err.Error())
		}

		flagTemplatePath = filepath.Join(filepath.Dir(execDir), "templates")
	}

	fmt.Printf("%s Detecting framework for the project at %s%s\n", common.MsgTitle, flagPath, common.MsgReset)

	detector, err := Detect(flagPath)
	if err != nil {
		fmt.Println(common.MsgError, err.Error(), common.MsgReset)
		return
	}
	analyzer := detector.Analyzer(flagPath, flagEnvironment)

	err = packs.Analyze(analyzer)
	if err != nil {
		fmt.Printf("%s Failed to analyze the project due to %s", common.MsgError, err.Error())
	}

	dockerfileWriter := DockerfileWriter{TemplateDir: flagTemplatePath, ShouldOverwrite: flagOverwrite}
	if err := dockerfileWriter.write(analyzer); err != nil {
		fmt.Printf("%s Failed to write Dockerfile due to %s\n", common.MsgError, err.Error())
	}

	serviceYAMLWriter := ServiceYAMLWriter{TemplateDir: flagTemplatePath, OutputDir: analyzer.GetRootDir(), ShouldOverwrite: flagOverwrite}
	serviceYAMLContext := NewServiceYAMLContext(analyzer)
	if err := serviceYAMLWriter.write(serviceYAMLContext); err != nil {
		fmt.Printf("%s Failed to write services.yml due to %s\n", common.MsgError, err.Error())
	}

	if len(analyzer.GetContext().Messages) > 0 {
		fmt.Printf("%s Warnings: \n", common.MsgWarn)
		for _, m := range analyzer.GetContext().Messages {
			fmt.Printf(" %s %s\n", common.MsgWarn, m)
		}
	}

	fmt.Println(common.MsgTitle, "\n Done", common.MsgReset)
}
