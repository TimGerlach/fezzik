package fezzik_test

import (
	"fmt"
	"time"

	. "github.com/cloudfoundry-incubator/fezzik"
	"github.com/cloudfoundry-incubator/receptor"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func NewLightweightLRP(guid string, numInstances int) receptor.DesiredLRPCreateRequest {
	return receptor.DesiredLRPCreateRequest{
		ProcessGuid: guid,
		Domain:      domain,
		RootFS:      rootFS,
		Instances:   numInstances,
		Setup: &models.DownloadAction{
			From:     "http://onsi-public.s3.amazonaws.com/grace.tar.gz",
			To:       "/tmp",
			CacheKey: "grace",
		},
		Action: &models.RunAction{
			Path: "/tmp/grace",
		},
		Monitor: &models.RunAction{
			Path: "nc",
			Args: []string{"-z", "127.0.0.1", "8080"},
		},
		Ports:    []uint16{8080},
		DiskMB:   128,
		MemoryMB: 64,
	}
}

func ActualLRPFetcher(processGuid string) func() ([]receptor.ActualLRPResponse, error) {
	return func() ([]receptor.ActualLRPResponse, error) {
		return client.ActualLRPsByProcessGuid(processGuid)
	}
}

var _ = Describe("Starting up a DesiredLRP", func() {
	for _, factor := range []int{1, 5, 10, 20, 40} {
		factor := factor

		Context(fmt.Sprintf("Starting up numCellx%d instances", factor), func() {
			var desiredLRP receptor.DesiredLRPCreateRequest
			var lrpReporter *LRPReporter
			var numInstances int

			BeforeEach(func() {
				numInstances = factor * numCells

				desiredLRP = NewLightweightLRP(guid, numInstances)
				Ω(client.CreateDesiredLRP(desiredLRP)).Should(Succeed())

				cells, err := client.Cells()
				Ω(err).ShouldNot(HaveOccurred())

				reportName := fmt.Sprintf("Running %d Instances Across %d Cells", numInstances, numCells)
				lrpReporter = NewLRPReporter(reportName, numInstances, cells)
			})

			AfterEach(func() {
				t := time.Now()
				client.DeleteDesiredLRP(desiredLRP.ProcessGuid)
				Eventually(ActualLRPFetcher(desiredLRP.ProcessGuid), 240).Should(BeEmpty())
				lrpReporter.EmitSummary()
				fmt.Printf("Time to delete:%s\n", time.Since(t))
				lrpReporter.Save()
			})

			It(fmt.Sprintf("should handle numCellx%d LRP instances", factor), func() {
				t := time.Now()
				for {
					Ω(time.Since(t)).Should(BeNumerically("<", 5*time.Minute), "timed out waiting for everything to come up!")
					actuals, err := client.ActualLRPsByProcessGuid(desiredLRP.ProcessGuid)
					Ω(err).ShouldNot(HaveOccurred())
					done := lrpReporter.ProcessActuals(actuals)
					if done {
						return
					}
					time.Sleep(200 * time.Millisecond)
				}
			})
		})
	}
})
