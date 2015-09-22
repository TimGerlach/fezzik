package fezzik

import (
	. "github.com/onsi/gomega"

	"github.com/nu7hatch/gouuid"
)

func NewGuid(prefix string) string {
	u, err := uuid.NewV4()
	Expect(err).NotTo(HaveOccurred())
	return prefix + "-" + u.String()
}
