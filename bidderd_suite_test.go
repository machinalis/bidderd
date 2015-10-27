package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBidderd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bidderd Suite")
}
