package fezzik

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/onsi/gomega"
)

type TaskReporter struct {
	ReportTime     time.Time
	ReportName     string
	NumCells       int
	TaskGuids      []string
	TimeToCreate   map[string]time.Duration
	TimeToComplete map[string]time.Duration
	FailedTasks    map[string]string

	lock *sync.Mutex
}

func NewTaskReporter(reportName string, numCells int, tasks []receptor.TaskCreateRequest) *TaskReporter {
	guids := []string{}

	for _, task := range tasks {
		guids = append(guids, task.TaskGuid)
	}

	return &TaskReporter{
		ReportTime:     time.Now(),
		ReportName:     reportName,
		NumCells:       numCells,
		TaskGuids:      guids,
		TimeToCreate:   map[string]time.Duration{},
		TimeToComplete: map[string]time.Duration{},
		FailedTasks:    map[string]string{},

		lock: &sync.Mutex{},
	}
}

func (r *TaskReporter) DidCreate(guid string) {
	dt := time.Since(r.ReportTime)
	r.lock.Lock()
	r.TimeToCreate[guid] = dt
	r.lock.Unlock()
}

func (r *TaskReporter) Completed(task receptor.TaskResponse) {
	//TODO
	//WHEN TASKS HAVE CELLIDs, INCLUDE THE CELLID SO WE NOW WHERE THIS WAS PLACED
	dt := time.Since(r.ReportTime)
	r.lock.Lock()
	r.TimeToComplete[task.TaskGuid] = dt
	if task.Failed {
		r.FailedTasks[task.TaskGuid] = task.FailureReason
	}
	r.lock.Unlock()
}

func (r *TaskReporter) EmitSummary() {
	fmt.Printf("\n%s\n%s\n", strings.Repeat("-", len(r.ReportName)), r.ReportName)

	numCompleted := len(r.TimeToComplete)
	numFailed := len(r.FailedTasks)
	numSucceeded := numCompleted - numFailed
	numRequested := len(r.TaskGuids)
	neverCompleted := numRequested - numCompleted
	fractionSucceeded := float64(numSucceeded) / float64(numRequested)
	fractionFailed := float64(numFailed) / float64(numRequested)
	fractionDidnotComplete := float64(neverCompleted) / float64(numRequested)

	fmt.Printf("Of %d Tasks:\n", numRequested)
	fmt.Printf("  %d (%.2f%%) Succeeded\n", numSucceeded, fractionSucceeded*100)
	if numFailed > 0 {
		fmt.Printf("  %d (%.2f%%) Failed\n", numFailed, fractionFailed*100)
		for guid, reason := range r.FailedTasks {
			fmt.Printf("    %s: %s\n", guid, reason)
		}
	}
	if neverCompleted > 0 {
		fmt.Printf("  %d (%.2f%%) Never Completed\n", neverCompleted, fractionDidnotComplete*100)
	}

	creationStats := DurationMapStats(r.TimeToCreate)
	PrintStatsReport("Creation time stats (in seconds)", creationStats)
	completionStats := DurationMapStats(r.TimeToComplete)
	PrintStatsReport("Completion time stats (in seconds)", completionStats)

}

func (r *TaskReporter) Save() {
	f, err := os.OpenFile("./reports.json", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	Ω(err).ShouldNot(HaveOccurred())

	_, err = f.WriteString("TASK_REPORT\n")
	Ω(err).ShouldNot(HaveOccurred())

	Ω(json.NewEncoder(f).Encode(r)).Should(Succeed())
	Ω(f.Close()).Should(Succeed())
}
