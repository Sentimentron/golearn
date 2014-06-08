package edf

const (
	// File format version
	EDF_VERSION = 1
	// Sets the maximum length of the mapping in bytes
	// Currently set (arbitrarily to 32 GiB)
	EDF_LENGTH = 32 * (1024 * 1024 * 1024)
)

const (
	// File will only be read, modifications fail
	EDF_READ_ONLY = iota
	// File will be read and written
	EDF_READ_WRITE
	// File will be created and opened with RW
	EDF_CREATE
)

const (
	// Don't call Sync before unmapping
	EDF_UNMAP_NOSYNC = iota
	EDF_UNMAP_SYNC
)
