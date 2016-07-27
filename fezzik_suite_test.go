package fezzik_test

import (
	"flag"
	"fmt"
	"net/url"
	"runtime"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/fezzik"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/say"

	"testing"
	"time"
)

var bbsAddress string
var bbsCA string
var bbsClientCert string
var bbsClientKey string
var consulAddress string

var publiclyAccessibleIP string
var numCells int

var bbsClient bbs.Client
var logger lager.Logger
var domain, rootFS, guid string
var startTime time.Time

func init() {
	flag.StringVar(&bbsAddress, "bbs-address", "http://10.244.16.130:8889", "http address for the bbs (required)")
	flag.StringVar(&bbsCA, "bbs-ca", "", "bbs ca cert")
	flag.StringVar(&bbsClientCert, "bbs-client-cert", "", "bbs client ssl certificate")
	flag.StringVar(&bbsClientKey, "bbs-client-key", "", "bbs client ssl key")
	flag.StringVar(&consulAddress, "consul-address", "http://127.0.0.1:8500", "http address for the consul agent (required)")
	flag.StringVar(&publiclyAccessibleIP, "publicly-accessible-ip", "10.0.2.2", "a publicly accessible IP for the host the test is running on (necssary to run a local server that containers can phone home to)")
	flag.IntVar(&numCells, "num-cells", 0, "number of cells")
	flag.Parse()
}

func TestFezzik(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fezzik Suite")
}

var _ = BeforeSuite(func() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	bbsClient = initializeBBSClient()

	logger = lagertest.NewTestLogger("fezzik")
	domain = "fezzik"
	rootFS = "preloaded:cflinuxfs2"

	if numCells == 0 {
		cells, err := bbsClient.Cells(logger)
		Expect(err).NotTo(HaveOccurred())
		numCells = len(cells)
	}

	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)

	say.Println(0, say.Green("Running Fezzik scaled to %d Cells", numCells))
})

var _ = BeforeEach(func() {
	startTime = time.Now()
	guid = fezzik.NewGuid(fmt.Sprintf("%s-%d", domain, GinkgoParallelNode()))
})

var _ = AfterEach(func() {
	endTime := time.Now()
	fmt.Fprint(
		GinkgoWriter,
		say.Cyan(
			"\n%s\nThis test referenced GUID %s\nStart time: %s (%d)\nEnd time: %s (%d)\n",
			CurrentGinkgoTestDescription().FullTestText,
			guid,
			startTime,
			startTime.Unix(),
			endTime,
			endTime.Unix(),
		),
	)
})

func initializeBBSClient() bbs.Client {
	bbsURL, err := url.Parse(bbsAddress)
	Expect(err).NotTo(HaveOccurred())

	if bbsURL.Scheme != "https" {
		return bbs.NewClient(bbsAddress)
	}

	bbsClient, err := bbs.NewSecureClient(bbsAddress, bbsCA, bbsClientCert, bbsClientKey, 0, 0)
	Expect(err).NotTo(HaveOccurred())
	return bbsClient
}
