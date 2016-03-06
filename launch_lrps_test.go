package fezzik_test

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/bbs/models"
	. "github.com/cloudfoundry-incubator/fezzik"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func NewLightweightLRP(guid string, numInstances int32) *models.DesiredLRP {
	return &models.DesiredLRP{
		ProcessGuid: guid,
		Domain:      domain,
		RootFs:      rootFS,
		Instances:   numInstances,
		Setup: models.WrapAction(&models.RunAction{
			Path: `{"duration_in_seconds":0}`,
			User: "vcap",
		}),
		Action: models.WrapAction(&models.RunAction{
			Path: `{"duration_in_seconds":1000}`,
			User: "vcap",
		}),
		Monitor: models.WrapAction(&models.RunAction{
			Path: `{"duration_in_seconds":0}`,
			User: "vcap",
		}),
		Ports:    []uint32{8080},
		DiskMb:   128,
		MemoryMb: 64,
	}
}

func ActualLRPFetcher(processGuid string) func() ([]*models.ActualLRPGroup, error) {
	return func() ([]*models.ActualLRPGroup, error) {
		return bbsClient.ActualLRPGroupsByProcessGuid(processGuid)
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
				By("Desiring DesiredLRP")
				Expect(bbsClient.DesireLRP(desiredLRP)).To(Succeed())

				cells, err := locketClient.Cells()
				Expect(err).NotTo(HaveOccurred())

				reportName := fmt.Sprintf("Running %d Instances Across %d Cells", numInstances, numCells)
				lrpReporter = NewLRPReporter(reportName, int(numInstances), cells)
			})

			AfterEach(func() {
				lrpReporter.EmitSummary()
				lrpReporter.Save()

				t := time.Now()
				By("Removing DesiredLRP")
				bbsClient.RemoveDesiredLRP((desiredLRP.ProcessGuid))
				Eventually(ActualLRPFetcher(desiredLRP.ProcessGuid), 240).Should(BeEmpty())
				fmt.Printf("Time to delete:%s\n", time.Since(t))
			})

			It(fmt.Sprintf("should handle numCellx%d LRP instances", factor), func() {
				t := time.Now()
				for {
					Expect(time.Since(t)).To(BeNumerically("<", 5*time.Minute), "timed out waiting for everything to come up!")
					actuals, err := bbsClient.ActualLRPGroupsByProcessGuid(desiredLRP.ProcessGuid)
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
