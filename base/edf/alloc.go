package edf

import (
	"fmt"
)

// ContentEntry structs are stored in ContentEntry blocks
// starting always at block 2
type ContentEntry struct {
	// Which thread this entry is assigned to
	Thread uint32
	// Which page this block starts at
	Start uint32
	// The page up to and including which the block ends
	End uint32
}

func (e *EdfFile) extend(additionalBytes uint32) error {
	fileInfo, err := e.f.Stat()
	if err != nil {
		panic(err)
	}
	newSize := fileInfo.Size() + int64(additionalBytes)
	return e.truncate(newSize)
}

func (e *EdfFile) getFreeMapSize() uint64 {
	fileInfo, err := e.f.Stat()
	if err != nil {
		panic(err)
	}
	return uint64(fileInfo.Size()) / e.pageSize
}

// FixedAlloc allocates a |bytesRequested| chunk of pages
// on the FIXED thread
func (e *EdfFile) FixedAlloc(bytesRequested uint32) (EdfRange, error) {
	pageSize := uint32(e.pageSize)
	return e.AllocPages((pageSize*bytesRequested+pageSize/2)/pageSize, 2)
}

func (e *EdfFile) getContiguousOffset(pagesRequested uint32) (uint32, error) {
	// Create the free bitmap
	bitmap := make([]bool, e.getFreeMapSize())
	bitmap[0] = true
	bitmap[1] = true
	// Traverse the contents table and build a free bitmap
	block := uint64(2)
	for {
		// Get the range for this block
		r := e.GetPageRange(block, block)
		if r.SegmentStart != r.SegmentEnd {
			return 0, fmt.Errorf("Contents block split across segments")
		}
		bytes := e.m[r.SegmentStart]
		bytes = bytes[r.ByteStart : r.ByteEnd+1]
		// Get the address of the next contents block
		block = uint64FromBytes(bytes)
		if block != 0 {
			// No point in checking this block for free space
			continue
		}
		bytes = bytes[8:]
		// Look for a blank entry in the table
		for i := 0; i < len(bytes); i += 12 {
			threadId := uint32FromBytes(bytes[i:])
			if threadId == 0 {
				continue
			}
			start := uint32FromBytes(bytes[i+4:])
			end := uint32FromBytes(bytes[i+8:])
			for j := start; j <= end; j++ {
				bitmap[j] = true
			}
		}
	}
	// Look through the freemap and find a good spot
	for i := 0; i < len(bitmap); i++ {
		if bitmap[i] {
			continue
		}
		for j := i; j < len(bitmap); j++ {
			if bitmap[j] {
				diff := j - 1 - i
				if diff > int(pagesRequested) {
					return uint32(i), nil
				}
			}
		}
	}
	return 0, nil
}

// addNewContentsBlock adds a new contents block in the next available space
func (e *EdfFile) addNewContentsBlock() error {

	var toc ContentEntry

	// Find the next available offset
	startBlock, err := e.getContiguousOffset(1)
	if startBlock == 0 && err == nil {
		// Increase the size of the file if necessary
		e.extend(uint32(e.pageSize))
	} else if err != nil {
		return err
	}

	// Traverse the contents blocks looking for one with a blank NEXT pointer
	block := uint64(0)
	for {
		// Get the range for this block
		r := e.GetPageRange(block, block)
		if r.SegmentStart != r.SegmentEnd {
			return fmt.Errorf("Contents block split across segments")
		}
		bytes := e.m[r.SegmentStart]
		bytes = bytes[r.ByteStart : r.ByteEnd+1]
		// Get the address of the next contents block
		block = uint64FromBytes(bytes)
		if block == 0 {
			uint64ToBytes(uint64(startBlock), bytes)
		}
	}
	// Add to the next available TOC space
	toc.Start = startBlock
	toc.End = startBlock + 1
	toc.Thread = 1 // SYSTEM thread
	return e.addToTOC(&toc, false)
}

// addToTOC adds a ContentsEntry structure in the next available place
func (e *EdfFile) addToTOC(c *ContentEntry, extend bool) error {
	// Traverse the contents table looking for a free spot
	block := uint64(2)
	for {
		// Get the range for this block
		r := e.GetPageRange(block, block)
		if r.SegmentStart != r.SegmentEnd {
			return fmt.Errorf("Contents block split across segments")
		}
		bytes := e.m[r.SegmentStart]
		bytes = bytes[r.ByteStart : r.ByteEnd+1]
		// Get the address of the next contents block
		block = uint64FromBytes(bytes)
		if block != 0 {
			// No point in checking this block for free space
			continue
		}
		bytes = bytes[8:]
		// Look for a blank entry in the table
		cur := 0
		for {
			threadId := uint32FromBytes(bytes)
			if threadId == 0 {
				break
			}
			cur += 12
			bytes = bytes[12:]
			if len(bytes) < 12 {
				if extend {
					// Append a new contents block and try again
					e.addNewContentsBlock()
				} else {
					return fmt.Errorf("Can't add to contents: no space available")
				}
			}
		}
		// Write the contents information into this block
		uint32ToBytes(c.Thread, bytes)
		bytes = bytes[4:]
		uint32ToBytes(c.Start, bytes)
		bytes = bytes[4:]
		uint32ToBytes(c.End, bytes)
		break
	}
	return nil
}

// FixedAllocPages allocates a |pagesRequested| chunk of pages
// on the FIXED thread
func (e *EdfFile) AllocPages(pagesRequested uint32, thread uint32) (EdfRange, error) {

	var ret EdfRange
	var toc ContentEntry

	// Find the next available offset
	startBlock, err := e.getContiguousOffset(pagesRequested)
	if startBlock == 0 && err == nil {
		// Increase the size of the file if necessary
		e.extend(pagesRequested * uint32(e.pageSize))
		return e.AllocPages(pagesRequested, thread)
	} else if err != nil {
		return ret, err
	}

	// Add to the table of contents
	toc.Thread = thread
	toc.Start = startBlock
	toc.End = startBlock + pagesRequested
	err = e.addToTOC(&toc, true)

	// Compute the range
	ret = e.GetPageRange(uint64(startBlock), uint64(pagesRequested))

	return ret, err
}
