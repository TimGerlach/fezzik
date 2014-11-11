package fezzik

import (
	. "github.com/onsi/gomega"

	"github.com/nu7hatch/gouuid"
)

func NewGuid() string {
	u, err := uuid.NewV4()
	Ω(err).ShouldNot(HaveOccurred())
	return u.String()
}
