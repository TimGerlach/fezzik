package fezzik_test

import (
	"flag"
	"log"
	"runtime"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/say"

	"testing"
	"time"
)

var receptorAddress, publiclyAccessibleIP string
var numCells int

var client receptor.Client
var domain, rootFS string

func init() {
	flag.StringVar(&receptorAddress, "receptor-address", "http://receptor.10.244.0.34.xip.io", "http address for the receptor (required)")
	flag.StringVar(&publiclyAccessibleIP, "publicly-accessible-ip", "10.0.2.2", "a publicly accessible IP for the host the test is running on (necssary to run a local server that containers can phone home to)")
	flag.IntVar(&numCells, "num-cells", 0, "number of cells")
	flag.Parse()

	if receptorAddress == "" {
		log.Fatal("i need a receptor-address to talk to Diego...")
	}
}

func TestFezzik(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fezzik Suite")
}

var _ = BeforeSuite(func() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	client = receptor.NewClient(receptorAddress)
	domain = "fezzik"
	rootFS = "preloaded:cflinuxfs2"

	if numCells == 0 {
		cells, err := client.Cells()
		Ω(err).ShouldNot(HaveOccurred())
		numCells = len(cells)
	}

	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)

	say.Println(0, say.Green("Running Fezzik scaled to %d Cells", numCells))
})
