package importx_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/xchapter7x/govmomi/govc/importx"
)

var _ = Describe("Archives", func() {
	var tmpDir string
	BeforeEach(func() {
		tmpDir, _ = ioutil.TempDir("/tmp", "importx_archives")
		os.Setenv("TMPDIR", tmpDir)
	})
	AfterEach(func() {
		os.Remove(tmpDir)
	})
	Describe("HTTPFileArchive struct", func() {
		var httpFileArchive *HTTPFileArchive

		Describe("given a Open() method", func() {
			Context("when called with a valid url string for a remote file", func() {
				var (
					fileReadCloser io.ReadCloser
					fileControl    = []byte(`hello there`)
					size           int64
					err            error
				)
				BeforeEach(func() {
					httpFileArchive = new(HTTPFileArchive)
					httpFileArchive.Client = &FakeClient{
						ResponseBody: fileControl,
					}
					fileReadCloser, size, err = httpFileArchive.Open("http://somefile.com")
				})

				It("then it should not return an error", func() {
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("then it should return the file's size ", func() {
					Ω(size).Should(Equal(int64(len(fileControl))))
				})

				It("then it should return the response body as a readcloser", func() {
					fileContents, _ := ioutil.ReadAll(fileReadCloser)
					Ω(fileContents).Should(Equal(fileControl))
				})
			})
		})
	})
})

type FakeClient struct {
	ResponseBody []byte
	DoCalled     int
	ErrFake      error
}

func (s *FakeClient) Do(req *http.Request) (res *http.Response, err error) {
	s.DoCalled++
	res = new(http.Response)
	bodyReader := bytes.NewReader(s.ResponseBody)
	res.Body = ioutil.NopCloser(bodyReader)
	res.ContentLength = int64(len(s.ResponseBody))
	return res, s.ErrFake
}
