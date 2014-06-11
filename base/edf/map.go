package edf

import (
	"fmt"
	mmap "github.com/riobard/go-mmap"
	"os"
)

// EdfFile represents a mapped file on disk or
// and anonymous mapping for instance storage
type EdfFile struct {
	f           *os.File
	m           []mmap.Mmap
	segmentSize uint
	pageSize    uint
}

// EdfRange represents a start and an end segment
// mapped in an EdfFile and also the byte offsets
// within that segment
type EdfRange struct {
	SegmentStart uint
	SegmentEnd   uint
	ByteStart    uint
	ByteEnd      uint
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

	// Set up various things
	ret.f = f
	ret.m = make([]mmap.Mmap, 0)

	// Figure out the flags
	protFlags := mmap.PROT_READ
	if mode == EDF_READ_WRITE || mode == EDF_CREATE {
		protFlags |= mmap.PROT_WRITE
	}
	mapFlags := mmap.MAP_FILE | mmap.MAP_SHARED

	// Get the page size
	pageSize := int64(os.Getpagesize())
	// Segment size is the size of each mapped region
	ret.pageSize = uint(pageSize)
	ret.segmentSize = uint(EDF_LENGTH) * uint(os.Getpagesize())

	// Map the file
	for i := int64(0); i < EDF_SIZE; i += int64(EDF_LENGTH) * pageSize {
		thisMapping, err := mmap.Map(f, i*pageSize, int(int64(EDF_LENGTH)*pageSize), protFlags, mapFlags)
		if err != nil {
			// TODO: cleanup
			return nil, err
		}
		ret.m = append(ret.m, thisMapping)
	}

	// Verify or generate the header
	if mode == EDF_READ_WRITE || mode == EDF_READ_ONLY {
		err = ret.VerifyHeader()
		if err != nil {
			return nil, err
		}
	} else if mode == EDF_CREATE {
		err = ret.truncate(4)
		if err != nil {
			return nil, err
		}
		ret.createHeader()
	} else {
		err = fmt.Errorf("Unrecognised flags")
	}

	return &ret, err
}

// GetRange returns the segment offset and range of
// two positions in the file
func (e *EdfFile) Range(byteStart uint, byteEnd uint) EdfRange {
	var ret EdfRange
	ret.SegmentStart = byteStart / e.segmentSize
	ret.SegmentEnd = byteEnd / e.segmentSize
	ret.ByteStart = byteStart % e.segmentSize
	ret.ByteEnd = byteEnd % e.segmentSize
	return ret
}

// VerifyHeader checks that this version of GoLearn can
// read the file presented
func (e *EdfFile) VerifyHeader() error {
	// Check the magic bytes
	diff := (e.m[0][0] ^ byte('G')) | (e.m[0][1] ^ byte('O'))
	diff |= (e.m[0][2] ^ byte('L')) | (e.m[0][3] ^ byte('N'))
	if diff != 0 {
		return fmt.Errorf("Invalid magic bytes")
	}
	// Check the file version
	version := uint32FromBytes(e.m[0][4:8])
	if version != EDF_VERSION {
		return fmt.Errorf("Unsupported version: %u", version)
	}
	// Check the page size
	pageSize := uint32FromBytes(e.m[0][8:12])
	if pageSize != uint32(os.Getpagesize()) {
		return fmt.Errorf("Unsupported page size: (file: %d, system: %d", pageSize, os.Getpagesize())
	}
	return nil
}

// createHeader writes a valid header file into the file.
// Unexported since it can cause data loss.
func (e *EdfFile) createHeader() {
	e.m[0][0] = byte('G')
	e.m[0][1] = byte('O')
	e.m[0][2] = byte('L')
	e.m[0][3] = byte('N')
	uint32ToBytes(EDF_VERSION, e.m[0][4:8])
	uint32ToBytes(uint32(os.Getpagesize()), e.m[0][8:12])
	e.Sync()
}

// writeInitialData writes system thread information
func (e *EdfFile) writeInitialData() {
	// Thread information goes in the second disk block
	// 8 bytes is left blank for the successor
	threadOffset := e.Range(e.pageSize + 8, (e.pageSize << 2) - 1)
	segment := e.m[threadOffset.SegmentStart]
	segment = segment[threadOffset.ByteStart:threadOffset.ByteEnd]
	// Write the declaration for the "system" thread which stores all of the
	// information on block allocations
	threadStr := "SYSTEM"
	threadLength := uint32(6)
	// Store the string length first
	uint32ToBytes(threadLength, segment)
	segment = segment[4:]
	// Copy the string
	copy(segment, threadStr)
	segment = segment[6:]
	// Write the location of the contents block
	// (0x2)
	uint32ToBytes(2, segment)
	// Write this thread number
	segment = segment[4:]
	uint32ToBytes(1, segment)
}

// Sync writes information to physical storage
func (e *EdfFile) Sync() error {
	for _, m := range e.m {
		err := m.Sync(mmap.MS_SYNC)
		if err != nil {
			return err
		}
	}
	return nil
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
	for _, m := range e.m {
		err := m.Unmap()
		if err != nil {
			return err
		}
	}
	return nil
}
