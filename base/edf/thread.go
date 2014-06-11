package edf

// Threads are streams of data encapsulated within the file
type Thread struct {
	name string
	id   uint32
}

// GetSpaceNeeded the number of bytes needed to serialize this
// Thread.
func (t *Thread) GetSpaceNeeded() int {
	return 8 + len(t.name)
}

// Serialize copies this thread to the output byte slice
// Returns the number of bytes used
func (t *Thread) Serialize(out []byte) int {
	// ret keeps track of written bytes
	ret := 0
	// Write the length of the name first
	nameLength := len(t.name)
	uint32ToBytes(uint32(nameLength), out)
	out = out[4:]
	ret += 4
	// Then write the string
	copy(out, t.name)
	out = out[nameLength:]
	ret += nameLength
	// Then the thread number
	uint32ToBytes(t.id, out)
	ret += 4
	return ret
}

// Deserialize copies the input byte slice into a thread
func (t *Thread) Deserialize(out []byte) int {
	ret := 0
	// Read the length of the thread's name
	nameLength := uint32FromBytes(out)
	ret += 4
	out = out[4:]
	// Copy out the string
	t.name = string(out[:nameLength])
	ret += int(nameLength)
	out = out[nameLength:]
	// Read the identifier
	t.id = uint32FromBytes(out)
	ret += 4
	return ret
}
