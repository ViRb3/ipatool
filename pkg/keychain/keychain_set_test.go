package keychain

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Keychain (Set)", func() {
	var (
		keychain Keychain
	)

	BeforeEach(func() {
		keychain = New(Args{
			Directory: GinkgoT().TempDir(),
		})
	})

	When("directory exists", func() {
		const testKey = "test-key"
		var testData = []byte("test")

		It("returns nil", func() {
			err := keychain.Set(testKey, testData)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("directory does not exist", func() {
		const testKey = "test-key"
		var testData = []byte("test")

		BeforeEach(func() {
			keychain = New(Args{
				Directory: filepath.Join(GinkgoT().TempDir(), "missing"),
			})
		})

		It("creates the directory", func() {
			err := keychain.Set(testKey, testData)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
