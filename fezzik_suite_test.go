package fezzik_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFezzik(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fezzik Suite")
}
