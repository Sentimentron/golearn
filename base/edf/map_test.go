package edf

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"
)

var (
	SegFault os.Signal = syscall.SIGSEGV
)

func TestCleanup(t *testing.T) {
	var ref []byte
	Convey("The mapping should be unlinked after garbage collection", t, func() {
		Convey("Anonymous mapping should suceed", func() {
			mapping, err := EdfAnonMap()
			So(err, ShouldEqual, nil)
			ref = mapping.m[0]
		})
		// Now letting the mapping fall out of scope
		// Forcing garbage collection twice
		runtime.GC()
		runtime.GC()
		Convey("Access the anonymous mapping should panic", func() {
			So(func() {
				ref[0] = 'H'
				ref[100] = 'E'
				ref[200] = 'L'
				ref[300] = 'L'
				ref[400] = 'O'
				ref[500] = 'W'
				ref[1000] = 'O'
				ref[2000] = 'R'
				ref[3000] = 'L'
				ref[4000] = 'D'
			}, ShouldPanic)
		})
	})
	Convey("Access to the mapping should fail after unmapping", t, func() {
		Convey("Anonymous mapping should suceed", func() {
			mapping, err := EdfAnonMap()
			So(err, ShouldEqual, nil)
			Convey("Unmap should succeed", func() {
				err := mapping.Unmap(EDF_UNMAP_NOSYNC)
				So(err, ShouldEqual, nil)
				Convey("Subsequent access should fail", func() {
					// Start swallowing segfaults
					c := make(chan os.Signal, 1)
					signal.Notify(c)
					faulted := false
					defer signal.Stop(c)

					mapping.m[0][500] = 2
					select {
					case s := <-c:
						if s == SegFault {
							faulted = true
						}

					case <-time.After(1 * time.Second):
						faulted = false
					}

					So(faulted, ShouldEqual, true)
				})
			})
		})
	})
}

func TestAnonMap(t *testing.T) {
	Convey("Anonymous mapping should suceed", t, func() {
		mapping, err := EdfAnonMap()
		So(err, ShouldEqual, nil)
		bytes := mapping.m[0]
		// Read the magic bytes
		magic := bytes[0:4]
		Convey("Magic bytes should be correct", func() {
			So(magic[0], ShouldEqual, byte('G'))
			So(magic[1], ShouldEqual, byte('O'))
			So(magic[2], ShouldEqual, byte('L'))
			So(magic[3], ShouldEqual, byte('N'))
		})
		// Read the file version
		versionBytes := bytes[4:8]
		Convey("Version should be correct", func() {
			version := uint32FromBytes(versionBytes)
			So(version, ShouldEqual, EDF_VERSION)
		})
		// Read the block size
		blockBytes := bytes[8:12]
		Convey("Page size should be correct", func() {
			pageSize := uint32FromBytes(blockBytes)
			So(pageSize, ShouldEqual, os.Getpagesize())
		})
	})
}

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
				version := uint32FromBytes(versionBytes)
				So(version, ShouldEqual, EDF_VERSION)
			})
			// Read the block size
			blockBytes := make([]byte, 4)
			read, err = tempFile.ReadAt(blockBytes, 8)
			Convey("Page size should be correct", func() {
				So(err, ShouldEqual, nil)
				So(read, ShouldEqual, 4)
				pageSize := uint32FromBytes(blockBytes)
				So(pageSize, ShouldEqual, os.Getpagesize())
			})
			// Check the file size is at least four * page size
			info, err := tempFile.Stat()
			Convey("File should be the right size", func() {
				So(err, ShouldEqual, nil)
				So(info.Size(), ShouldBeGreaterThanOrEqualTo, 4*os.Getpagesize())
			})
		})
	})
}

func TestFileThreadCounter(t *testing.T) {
	Convey("Creating a non-existent file should succeed", t, func() {
		tempFile, err := ioutil.TempFile(os.TempDir(), "TestFileCreate")
		So(err, ShouldEqual, nil)
		Convey("Mapping the file should suceed", func() {
			mapping, err := EdfMap(tempFile, EDF_CREATE)
			So(err, ShouldEqual, nil)
			Convey("The file should have two threads to start with", func() {
				count := mapping.GetThreadCount()
				So(count, ShouldEqual, 2)
				Convey("They should be SYSTEM and FIXED", func() {
					threads, err := mapping.GetThreads()
					So(err, ShouldEqual, nil)
					So(len(threads), ShouldEqual, 2)
					So(threads[1], ShouldEqual, "SYSTEM")
					So(threads[2], ShouldEqual, "FIXED")
				})
			})
			Convey("Incrementing the threadcount should result in three threads", func() {
				mapping.incrementThreadCount()
				count := mapping.GetThreadCount()
				So(count, ShouldEqual, 3)
				Convey("Thread information should indicate corruption", func() {
					_, err := mapping.GetThreads()
					So(err, ShouldNotEqual, nil)
				})
			})
		})
	})
}
