package fezzik_test

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/fezzik"
	"code.cloudfoundry.org/lager"
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
var skipVerifyCertificate bool
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
	flag.BoolVar(&skipVerifyCertificate, "bbs-skip-verify-certificate", false, "whether to ignore invalid TLS certificates")
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

	logger = lager.NewLogger("fezzik")
	sink := lager.NewWriterSink(os.Stdout, lager.ERROR)
	logger.RegisterSink(sink)
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

	bbsClient, err := getSecureClient(skipVerifyCertificate)

	Expect(err).NotTo(HaveOccurred())
	return bbsClient
}

func getSecureClient(skipVerify bool) (bbs.Client, error) {
	if skipVerifyCertificate {
		return bbs.NewSecureSkipVerifyClient(bbsAddress, bbsClientCert, bbsClientKey, 0, 0)
	} else {
		return bbs.NewSecureClient(bbsAddress, bbsCA, bbsClientCert, bbsClientKey, 0, 0)
	}
}
