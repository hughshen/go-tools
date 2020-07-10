package main

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	wd         string
	command    = flag.String("cmd", "", "Command")
	workDir    = flag.String("work", "", "Working directory")
	includeDir = flag.String("hasdir", "", "Include directories, split by comma")
	excludeDir = flag.String("notdir", "", "Exclude directories, split by comma")
	includeExt = flag.String("hasext", "", "Include extensions, split by comma")
	excludeExt = flag.String("notext", "", "Exclude extensions, split by comma")
)

func main() {
	flag.Parse()

	var err error
	if *workDir != "" {
		wd, err = filepath.Abs(*workDir)
	} else {
		wd, err = os.Getwd()
	}
	if err != nil {
		log.Fatal(err)
	}

	watch(runCmd())
}

func killCmd(cmd *exec.Cmd) error {
	if err := cmd.Process.Kill(); err != nil {
		return err
	}

	_, err := cmd.Process.Wait()
	return err
}

func runCmd() *exec.Cmd {
	var cmd *exec.Cmd

	if *command != "" {
		arr := strings.Split(*command, " ")
		cmd = exec.Command(arr[0], arr[1:]...)
	} else {
		// Go app
		_, appName := filepath.Split(wd)
		subCmd := exec.Command("go", "build")
		subCmd.Dir = wd
		_, err := subCmd.Output()
		if err != nil {
			switch err.(type) {
			case *exec.ExitError:
				log.Fatal(string(err.(*exec.ExitError).Stderr))
			default:
				log.Fatal(err)
			}
		}

		*excludeDir += "vendor"
		*excludeExt += "_test.go"
		*includeExt += ".go"

		cmd = exec.Command("./" + appName)
	}

	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	return cmd
}

func watch(cmd *exec.Cmd) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for event := range watcher.Events {
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("%c[1;40;32mmodified file: %s%c[0m\n", 0x1B, event.Name, 0x1B)
				if cmdErr := killCmd(cmd); cmdErr != nil {
					log.Fatal(cmdErr)
				}
				cmd = runCmd()
			}
		}
	}()

	errs := []error{}

	files := getFiles(wd)
	for _, p := range files {
		errs = append(errs, watcher.Add(p))
	}

	for _, err = range errs {
		if err != nil {
			log.Fatal(err)
		}
	}

	<-make(chan struct{})
}

func getFiles(path string) []string {
	results := []string{}

	folder, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer folder.Close()

	files, _ := folder.Readdir(-1)
	for _, file := range files {
		fileName := file.Name()
		newPath := path + "/" + fileName

		isValidDir := file.IsDir() && !strings.HasPrefix(fileName, ".") && checkDir(fileName)

		isValidFile := !file.IsDir() && checkExt(fileName)

		if isValidDir {
			results = append(results, getFiles(newPath)...)
		} else if isValidFile {
			results = append(results, newPath)
		}
	}

	return results
}

func checkFileName(fileName, check, checkType string) bool {
	for _, v := range strings.Split(check, ",") {
		if checkType == "prefix" {
			if strings.HasPrefix(fileName, v) {
				return true
			}
		} else if checkType == "suffix" {
			if strings.HasSuffix(fileName, v) {
				return true
			}
		} else if checkType == "equal" {
			if fileName == v {
				return true
			}
		}
	}

	return false
}

func checkDir(fileName string) bool {
	if *excludeDir != "" {

		if checkFileName(fileName, *excludeDir, "equal") == true {
			return false
		}
	}

	if *includeDir != "" {
		if checkFileName(fileName, *includeDir, "equal") == true {
			return true
		} else {
			return false
		}
	}

	return true
}

func checkExt(fileName string) bool {
	if *excludeExt != "" {
		if checkFileName(fileName, *excludeExt, "suffix") == true {
			return false
		}
	}

	if *includeExt != "" {
		if checkFileName(fileName, *includeExt, "suffix") == true {
			return true
		} else {
			return false
		}
	}

	return true
}
