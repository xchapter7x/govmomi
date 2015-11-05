package importx_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/xchapter7x/govmomi/govc/importx"
)

var _ = Describe("OVF", func() {
	Describe("given a GetArchive function", func() {
		var archive Archive
		Context("when called with a http remote file path", func() {
			BeforeEach(func() {
				archive = GetArchive("http://mysamplefile.com")
			})
			It("then we should return a HTTPFileArchive object", func() {
				Ω(archive).Should(BeAssignableToTypeOf(&HTTPFileArchive{}))
			})

			It("then we should return a HTTPFileArchive object", func() {
				Ω(archive).ShouldNot(BeAssignableToTypeOf(&FileArchive{}))
			})
		})

		Context("when called with a local file path", func() {
			BeforeEach(func() {
				archive = GetArchive("/tmp/mysamplefile.txt")
			})
			It("then we should return a FileArchive object", func() {
				Ω(archive).Should(BeAssignableToTypeOf(&FileArchive{}))
			})
			It("then we should return a FileArchive object", func() {
				Ω(archive).ShouldNot(BeAssignableToTypeOf(&HTTPFileArchive{}))
			})
		})
	})
})
