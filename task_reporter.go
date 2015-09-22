package fezzik

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/locket/presence"
	. "github.com/onsi/gomega"
	"github.com/onsi/say"
)

type TaskReporter struct {
	ReportTime       time.Time
	ReportName       string
	NumCells         int
	NumRequested     int
	TimeToCreate     map[string]time.Duration
	TimeToComplete   map[string]time.Duration
	FailedTasks      map[string]string
	TaskDistribution map[string]int

	lock *sync.Mutex
}

func NewTaskReporter(reportName string, numRequested int, cells []presence.CellPresence) *TaskReporter {
	taskDistribution := map[string]int{}
	for _, cell := range cells {
		taskDistribution[cell.CellID] = 0
	}

	return &TaskReporter{
		ReportTime:       time.Now(),
		ReportName:       reportName,
		NumRequested:     numRequested,
		NumCells:         len(cells),
		TimeToCreate:     map[string]time.Duration{},
		TimeToComplete:   map[string]time.Duration{},
		FailedTasks:      map[string]string{},
		TaskDistribution: taskDistribution,

		lock: &sync.Mutex{},
	}
}

func (r *TaskReporter) DidCreate(guid string) {
	dt := time.Since(r.ReportTime)
	r.lock.Lock()
	r.TimeToCreate[guid] = dt
	r.lock.Unlock()
}

func (r *TaskReporter) Completed(task *models.Task) {
	dt := time.Since(r.ReportTime)
	r.lock.Lock()
	r.TimeToComplete[task.TaskGuid] = dt
	r.TaskDistribution[task.CellId] += 1
	if task.Failed {
		r.FailedTasks[task.TaskGuid] = task.FailureReason
	}
	r.lock.Unlock()
}

func (r *TaskReporter) EmitSummary() {
	say.Println(0, "")
	say.Println(0, strings.Repeat("-", len(r.ReportName)))
	say.Println(0, r.ReportName)

	numCompleted := len(r.TimeToComplete)
	numFailed := len(r.FailedTasks)
	numSucceeded := numCompleted - numFailed
	numRequested := r.NumRequested
	neverCompleted := numRequested - numCompleted
	fractionSucceeded := float64(numSucceeded) / float64(numRequested)
	fractionFailed := float64(numFailed) / float64(numRequested)
	fractionDidnotComplete := float64(neverCompleted) / float64(numRequested)

	say.Println(0, "Of %d Tasks:", numRequested)
	say.Println(1, say.Green("%d (%.2f%%) Succeeded", numSucceeded, fractionSucceeded*100))
	if numFailed > 0 {
		say.Println(1, say.Red("%d (%.2f%%) Failed", numFailed, fractionFailed*100))
		for guid, reason := range r.FailedTasks {
			say.Println(2, "%s: %s", guid, say.Red(reason))
		}
	}
	if neverCompleted > 0 {
		say.Println(1, say.Red("%d (%.2f%%) Never Completed", neverCompleted, fractionDidnotComplete*100))
	}

	creationStats := DurationMapStats(r.TimeToCreate)
	PrintStatsReport("Creation time stats (in seconds)", creationStats)

	completionStats := DurationMapStats(r.TimeToComplete)
	PrintStatsReport("Completion time stats (in seconds)", completionStats)

	cells := []string{}
	for cell := range r.TaskDistribution {
		cells = append(cells, cell)
	}

	say.Println(0, "Distribution:")
	sort.Strings(cells)
	for _, cell := range cells {
		say.Println(1, "%12s %s", cell, say.Yellow(strings.Repeat("+", r.TaskDistribution[cell])))
	}
}

func (r *TaskReporter) Save() {
	f, err := os.OpenFile("./reports.json", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	Expect(err).NotTo(HaveOccurred())

	_, err = f.WriteString("TASK_REPORT\n")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewEncoder(f).Encode(r)).To(Succeed())
	Expect(f.Close()).To(Succeed())
}
