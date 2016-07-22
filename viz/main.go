package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/fezzik"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("point me at reports.json!")
		os.Exit(1)
	}

	taskReports, lrpReports, err := LoadReports(os.Args[1])
	if err != nil {
		fmt.Println("failed to load report\n", err.Error())
		os.Exit(1)
	}

	//make these plot...
	for _, task := range taskReports {
		task.EmitSummary()
	}
	for _, lrp := range lrpReports {
		lrp.EmitSummary()
	}
}

func LoadReports(filename string) ([]*fezzik.TaskReporter, []*fezzik.LRPReporter, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	taskReporters := []*fezzik.TaskReporter{}
	lrpReporters := []*fezzik.LRPReporter{}

	lines := strings.Split(string(content), "\n")
	for i := 0; i < len(lines)/2; i++ {
		typeIndex := i * 2
		contentIndex := typeIndex + 1
		switch lines[typeIndex] {
		case "TASK_REPORT":
			taskReport := &fezzik.TaskReporter{}
			err := json.Unmarshal([]byte(lines[contentIndex]), &taskReport)
			if err != nil {
				return nil, nil, err
			}
			taskReporters = append(taskReporters, taskReport)
		case "LRP_REPORT":
			lrpReport := &fezzik.LRPReporter{}
			err := json.Unmarshal([]byte(lines[contentIndex]), &lrpReport)
			if err != nil {
				return nil, nil, err
			}
			lrpReporters = append(lrpReporters, lrpReport)
		default:
			return nil, nil, fmt.Errorf("unkown report type: %s", lines[typeIndex])
		}
	}

	return taskReporters, lrpReporters, nil
}
