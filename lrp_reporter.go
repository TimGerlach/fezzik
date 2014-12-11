package fezzik

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/veritas/say"
)

type LRPReporter struct {
	ReportTime      time.Time
	ReportName      string
	NumCells        int
	NumInstances    int
	TimeToClaimed   map[string]time.Duration
	TimeToRunning   map[string]time.Duration
	LRPDistribution map[string]int

	lock *sync.Mutex
}

func NewLRPReporter(reportName string, numInstances int, cells []receptor.CellResponse) *LRPReporter {
	lrpDistribution := map[string]int{}
	for _, cell := range cells {
		lrpDistribution[cell.CellID] = 0
	}

	return &LRPReporter{
		ReportTime:      time.Now(),
		ReportName:      reportName,
		NumCells:        len(cells),
		NumInstances:    numInstances,
		TimeToClaimed:   map[string]time.Duration{},
		TimeToRunning:   map[string]time.Duration{},
		LRPDistribution: lrpDistribution,

		lock: &sync.Mutex{},
	}
}

func (r *LRPReporter) ProcessActuals(actuals []receptor.ActualLRPResponse) bool {
	dt := time.Since(r.ReportTime)
	n := 0
	r.lock.Lock()
	for _, actual := range actuals {
		index := fmt.Sprintf("%d", actual.Index)
		if actual.State == receptor.ActualLRPStateClaimed || actual.State == receptor.ActualLRPStateRunning {
			if _, ok := r.TimeToClaimed[index]; !ok {
				r.TimeToClaimed[index] = dt
				r.LRPDistribution[actual.CellID] += 1
			}
		}
		if actual.State == receptor.ActualLRPStateRunning {
			n += 1
			if _, ok := r.TimeToRunning[index]; !ok {
				r.TimeToRunning[index] = dt
			}
		}
	}
	r.lock.Unlock()

	return n == r.NumInstances
}

func (r *LRPReporter) EmitSummary() {
	say.Println(0, "")
	say.Println(0, strings.Repeat("-", len(r.ReportName)))
	say.Println(0, r.ReportName)

	claimStats := DurationMapStats(r.TimeToClaimed)
	PrintStatsReport("Claim time stats (in seconds)", claimStats)

	runningStats := DurationMapStats(r.TimeToRunning)
	PrintStatsReport("Running time stats (in seconds)", runningStats)

	cells := []string{}
	for cell := range r.LRPDistribution {
		cells = append(cells, cell)
	}

	say.Println(0, "Distribution:")
	sort.Strings(cells)
	for _, cell := range cells {
		say.Println(1, "%12s %s", cell, say.Yellow(strings.Repeat("+", r.LRPDistribution[cell])))
	}
}

func (r *LRPReporter) Save() {
	f, err := os.OpenFile("./reports.json", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	Ω(err).ShouldNot(HaveOccurred())

	_, err = f.WriteString("LRP_REPORT\n")
	Ω(err).ShouldNot(HaveOccurred())

	Ω(json.NewEncoder(f).Encode(r)).Should(Succeed())
	Ω(f.Close()).Should(Succeed())
}
