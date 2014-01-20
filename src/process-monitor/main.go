package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func GetPids(processes ...string) map[string]map[string]uint {
	results := map[string]map[string]uint{}

	dir, err := os.Open("/proc")

	if err != nil {
		fmt.Println("Could not open /proc for reading: " + err.Error())
		return nil
	}

	defer dir.Close()

	proc_files, err := dir.Readdirnames(0)

	if err != nil {
		fmt.Println("Could not read directory names from /proc: " + err.Error())
		return nil
	}

	all_pids := []string{}
	// XXX totally cheating here -- the only all-numeric filenames in this dir
	// will be pid directories. This should be faster than 4 bajillion stat
	// calls (that I'd have to do this to anyway).
	for _, fn := range proc_files {
		_, err := strconv.Atoi(fn)
		if err == nil {
			all_pids = append(all_pids, fn)
		}
	}

	for _, pid := range all_pids {
		path := "/proc/" + pid + "/exe"
		exe, err := os.Readlink(path)

		if err != nil && !os.IsNotExist(err) {
			fmt.Println("Could not open " + path + ". Are you root? error: " + err.Error())
			return nil
		}

		for _, process := range processes {
			if exe == process {
				dir, err := os.Open("/proc/" + pid + "/fd")

				if err != nil && !os.IsNotExist(err) {
					fmt.Println("Could not open " + path + ". Are you root? error: " + err.Error())
					return nil
				}

				if results[process] == nil {
					results[process] = make(map[string]uint)
				}
				results[process]["count"]++
				nms, _ := dir.Readdirnames(0)
				results[process]["fds"] += uint(len(nms))
			}
		}
	}

	return results
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("enter the full path to a binary")
		os.Exit(1)
	}

	content, _ := json.Marshal(GetPids(os.Args[1:]...))

	fmt.Println(string(content))
}
