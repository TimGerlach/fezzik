package fezzik_test

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/bbs/models"
	. "code.cloudfoundry.org/fezzik"
	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func NewLightweightLRP(guid string, numInstances int32) *models.DesiredLRP {
	return &models.DesiredLRP{
		ProcessGuid: guid,
		Domain:      domain,
		RootFs:      rootFS,
		Instances:   numInstances,
		Setup: models.WrapAction(&models.DownloadAction{
			From:     "http://onsi-public.s3.amazonaws.com/grace.tar.gz",
			To:       "/tmp",
			CacheKey: "grace",
			User:     "vcap",
		}),
		Action: models.WrapAction(&models.RunAction{
			Path: "/tmp/grace",
			User: "vcap",
		}),
		Monitor: models.WrapAction(&models.RunAction{
			Path: "nc",
			Args: []string{"-z", "127.0.0.1", "8080"},
			User: "vcap",
		}),
		Ports:    []uint32{8080},
		DiskMb:   128,
		MemoryMb: 64,
	}
}

func ActualLRPFetcher(logger lager.Logger, processGuid string) func() ([]*models.ActualLRPGroup, error) {
	return func() ([]*models.ActualLRPGroup, error) {
		return bbsClient.ActualLRPGroupsByProcessGuid(logger, processGuid)
	}
}

var _ = Describe("Starting up a DesiredLRP", func() {
	for _, factor := range []int{1, 5, 10, 20, 40} {
		factor := factor

		Context(fmt.Sprintf("Starting up numCellx%d instances", factor), func() {
			var desiredLRP *models.DesiredLRP
			var lrpReporter *LRPReporter
			var numInstances int32

			BeforeEach(func() {
				numInstances = int32(factor * numCells)

				desiredLRP = NewLightweightLRP(guid, numInstances)
				Expect(bbsClient.DesireLRP(logger, desiredLRP)).To(Succeed())

				cells, err := bbsClient.Cells(logger)
				Expect(err).NotTo(HaveOccurred())

				reportName := fmt.Sprintf("Running %d Instances Across %d Cells", numInstances, numCells)
				lrpReporter = NewLRPReporter(reportName, int(numInstances), cells)
			})

			AfterEach(func() {
				lrpReporter.EmitSummary()
				lrpReporter.Save()

				t := time.Now()
				bbsClient.RemoveDesiredLRP(logger, desiredLRP.ProcessGuid)
				Eventually(ActualLRPFetcher(logger, desiredLRP.ProcessGuid), 240).Should(BeEmpty())
				fmt.Printf("Time to delete:%s\n", time.Since(t))
			})

			It(fmt.Sprintf("should handle numCellx%d LRP instances", factor), func() {
				t := time.Now()
				for {
					Expect(time.Since(t)).To(BeNumerically("<", 5*time.Minute), "timed out waiting for everything to come up!")
					actuals, err := bbsClient.ActualLRPGroupsByProcessGuid(logger, desiredLRP.ProcessGuid)
					Expect(err).NotTo(HaveOccurred())
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
