package edf

import (
	"fmt"
	mmap "github.com/riobard/go-mmap"
	"os"
)

// EdfFile represents a mapped file on disk or
// and anonymous mapping for instance storage
type EdfFile struct {
	f *os.File
	m mmap.Mmap
}

// EdfMap takes an os.File and returns an EdfMappedFile
// structure, which represents the mmap'd underlying file
//
// The `mode` parameter takes the following values
// 	EDF_CREATE EdfMap will truncate the file to the right length
//          and write the correct header information
//      EDF_READ_WRITE EdfMap will verify header information
//	EDF_READ_ONLY  EdfMap will verify header information
// IMPORTANT: EDF_LENGTH (edf.go) controls the size of the address
// space mapping. This means that the file can be truncated to the
// correct size without remapping. On 32-bit systems, this
// is set to 2GiB.
func EdfMap(f *os.File, mode int) (*EdfFile, error) {
	var ret EdfFile
	var err error

	// Assign os.File
	ret.f = f

	// Figure out the flags
	protFlags := mmap.PROT_READ
	if mode == EDF_READ_WRITE || mode == EDF_CREATE {
		protFlags |= mmap.PROT_WRITE
	}
	mapFlags := mmap.MAP_FILE | mmap.MAP_SHARED

	// Detect the current architecture
	maxLength := 64 * (1024 * 1024)
	/*	maxLength := 2 * (1000 * 1000 * 1000)
		if runtime.GOARCH == "amd64" {
		maxLength = EDF_LENGTH
	}*/

	// Map the file
	ret.m, err = mmap.Map(f, 0, maxLength, protFlags, mapFlags)
	if err != nil {
		return nil, err
	}

	// Verify or generate the header
	if mode == EDF_READ_WRITE || mode == EDF_READ_ONLY {
		err = ret.VerifyHeader()
		if err != nil {
			return nil, err
		}
	} else if mode == EDF_CREATE {
		err = ret.truncate(2)
		if err != nil {
			return nil, err
		}
		ret.createHeader()
	} else {
		err = fmt.Errorf("Unrecognised flags")
	}

	return &ret, err
}

// VerifyHeader checks that this version of GoLearn can
// read the file presented
func (e *EdfFile) VerifyHeader() error {
	// Check the magic bytes
	diff := (e.m[0] ^ byte('G')) | (e.m[1] ^ byte('O'))
	diff |= (e.m[2] ^ byte('L')) | (e.m[3] ^ byte('N'))
	if diff != 0 {
		return fmt.Errorf("Invalid magic bytes")
	}
	// Check the file version
	version := int32FromBytes(e.m[4:8])
	if version != EDF_VERSION {
		return fmt.Errorf("Unsupported version: %u", version)
	}
	// Check the page size
	pageSize := int32FromBytes(e.m[8:12])
	if pageSize != int32(os.Getpagesize()) {
		return fmt.Errorf("Unsupported page size: (file: %d, system: %d", pageSize, os.Getpagesize())
	}
	return nil
}

// createHeader writes a valid header file into the file.
// Unexported since it can cause data loss.
func (e *EdfFile) createHeader() {
	e.m[0] = byte('G')
	e.m[1] = byte('O')
	e.m[2] = byte('L')
	e.m[3] = byte('N')
	int32ToBytes(EDF_VERSION, e.m[4:8])
	int32ToBytes(int32(os.Getpagesize()), e.m[8:12])
	e.Sync()
}

// Sync writes information to physical storage
func (e *EdfFile) Sync() error {
	return e.m.Sync(mmap.MS_SYNC)
}

// truncate changes the size of the underlying file
// The size of the address space doesn't change.
func (e *EdfFile) truncate(size int64) error {
	pageSize := int64(os.Getpagesize())
	newSize := pageSize * size

	// Synchronise
	e.Sync()

	// Double-check that we're not reducing file size
	fileInfo, err := e.f.Stat()
	if err != nil {
		return err
	}
	if fileInfo.Size() > newSize {
		return fmt.Errorf("Can't reduce file size!")
	}

	// Truncate the file
	err = e.f.Truncate(newSize)
	if err != nil {
		return err
	}

	// Verify that the file is larger now than it was
	fileInfo, err = e.f.Stat()
	if err != nil {
		return err
	}
	if fileInfo.Size() != newSize {
		return fmt.Errorf("Truncation failed: %d, %d", fileInfo.Size(), newSize)
	}
	return err
}

// Unmap unlinks the EdfFile from the address space.
// EDF_UNMAP_NOSYNC skips calling Sync() on the underlying
// file before this happens.
// IMPORTANT: attempts to use this mapping after Unmap() is
// called will result in crashes.
func (e *EdfFile) Unmap(flags int) error {
	// Sync the file
	e.Sync()
	if flags != EDF_UNMAP_NOSYNC {
		e.Sync()
	}
	// Unmap the file
	// 	err := e.m.Unmap()
	return nil
}
