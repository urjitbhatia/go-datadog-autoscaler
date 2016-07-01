package scaler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestScaler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "go-datadog-scaler: Scaler test Suite")
}
