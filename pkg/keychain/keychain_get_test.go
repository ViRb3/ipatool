package keychain

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Keychain (Get)", func() {
	var (
		keychain Keychain
		dir      string
	)

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		keychain = New(Args{
			Directory: dir,
		})
	})

	When("item does not exist", func() {
		const testKey = "test-key"

		It("returns wrapped error", func() {
			data, err := keychain.Get(testKey)
			Expect(err).To(HaveOccurred())
			Expect(data).To(BeNil())
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
			data, err := keychain.Get(testKey)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(testData))
		})
	})

	When("item is invalid", func() {
		const testKey = "test-key"

		BeforeEach(func() {
			err := os.WriteFile(filepath.Join(dir, testKey), []byte("invalid"), 0600)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns wrapped error", func() {
			data, err := keychain.Get(testKey)
			Expect(err).To(HaveOccurred())
			Expect(data).To(BeNil())
		})
	})
})
