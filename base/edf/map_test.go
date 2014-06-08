package edf

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileCreate(t *testing.T) {
	Convey("Creating a non-existent file should succeed", t, func() {
		tempFile, err := ioutil.TempFile(os.TempDir(), "TestFileCreate")
		So(err, ShouldEqual, nil)
		Convey("Mapping the file should suceed", func() {
			mapping, err := EdfMap(tempFile, EDF_CREATE)
			So(err, ShouldEqual, nil)
			Convey("Unmapping the file should suceed", func() {
				err = mapping.Unmap(EDF_UNMAP_SYNC)
				So(err, ShouldEqual, nil)
			})

			// Read the magic bytes
			magic := make([]byte, 4)
			read, err := tempFile.ReadAt(magic, 0)
			Convey("Magic bytes should be correct", func() {
				So(err, ShouldEqual, nil)
				So(read, ShouldEqual, 4)
				So(magic[0], ShouldEqual, byte('G'))
				So(magic[1], ShouldEqual, byte('O'))
				So(magic[2], ShouldEqual, byte('L'))
				So(magic[3], ShouldEqual, byte('N'))
			})
			// Read the file version
			versionBytes := make([]byte, 4)
			read, err = tempFile.ReadAt(versionBytes, 4)
			Convey("Version should be correct", func() {
				So(err, ShouldEqual, nil)
				So(read, ShouldEqual, 4)
				version := int32FromBytes(versionBytes)
				So(version, ShouldEqual, EDF_VERSION)
			})
			// Read the block size
			blockBytes := make([]byte, 4)
			read, err = tempFile.ReadAt(blockBytes, 8)
			Convey("Page size should be correct", func() {
				So(err, ShouldEqual, nil)
				So(read, ShouldEqual, 4)
				pageSize := int32FromBytes(blockBytes)
				So(pageSize, ShouldEqual, os.Getpagesize())
			})
			// Check the file size is two * page size
			info, err := tempFile.Stat()
			Convey("File should be the right size", func() {
				So(err, ShouldEqual, nil)
				So(info.Size(), ShouldEqual, 2*os.Getpagesize())
			})
		})
	})
}
