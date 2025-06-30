package keychain

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Keychain (Remove)", func() {
	var (
		keychain Keychain
	)

	BeforeEach(func() {
		keychain = New(Args{
			Directory: GinkgoT().TempDir(),
		})
	})

	When("item does not exist", func() {
		const testKey = "test-key"

		It("returns wrapped error", func() {
			err := keychain.Remove(testKey)
			Expect(err).To(HaveOccurred())
		})
	})

	When("item exists", func() {
		const testKey = "test-key"
		var testData = []byte("test")

		BeforeEach(func() {
			err := keychain.Set(testKey, testData)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns data", func() {
			err := keychain.Remove(testKey)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
