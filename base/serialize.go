package base

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	//"reflect"
)

const (
	SerializationFormatVersion = "golearn 0.5"
)

func SerializeInstancesToFile(inst FixedDataGrid, path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	err = SerializeInstances(inst, f)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("Couldn't flush file: %s", err)
	}
	f.Close()
	return nil
}

func writeAttributesToFilePart(attrs []Attribute, f *tar.Writer, name string) error {
	// Get the marshaled Attribute array
	body, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	// Write a header
	hdr := &tar.Header{
		Name: name,
		Size: int64(len(body)),
	}
	if err := f.WriteHeader(hdr); err != nil {
		return err
	}

	// Write the marshaled data
	if _, err := f.Write([]byte(body)); err != nil {
		return err
	}

	return nil
}

func getTarContent(tr *tar.Reader, name string) []byte {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		if hdr.Name == name {
			ret := make([]byte, hdr.Size)
			n, err := tr.Read(ret)
			if int64(n) != hdr.Size {
				panic("Size mismatch")
			}
			if err != nil {
				panic(err)
			}
			return ret
		}
	}
	panic("File not found!")
}

func DeserializeInstances(f io.Reader) (FixedDataGrid, error) {
	// Define a JSON shim Attribute
	type JSONAttribute struct {
		Type string `json:type`
		Attr json.RawMessage
	}

	/*var attrs []JSONAttribute
	var specs []JSONAttribute

		// Open the .gz layer
		gzReader, err := gzip.NewReader(f)
		if err != nil {
			return nil, fmt.Errorf("Can't open: %s", err)
		}
		// Open the .tar layer
		tr := tar.NewReader(gzReader)
		// Retrieve the MANIFEST and verify
		manifestBytes := getTarContent(tr, "MANIFEST")
		if !reflect.DeepEqual(manifestBytes, []byte(SerializationFormatVersion)) {
			return nil, fmt.Errorf("Unsupported MANIFEST: %s", string(manifestBytes))
		}
		// Unmarshal the Attributes
		attrBytes := getTarContent(tr, "ATTRS")
		fmt.Println(string(attrBytes))
		err = json.Unmarshal(attrBytes, &attrs)
		if err != nil {
			return nil, fmt.Errorf("Attribute decode error: %s", err)
		}

		for _, a := range attrs {
			var attr Attribute
			var err error
			switch a.Type {
			case "binary":
				attr = new(BinaryAttribute)
				break
			case "float":
				attr = new(FloatAttribute)
				break
			case "categorical":
				attr = new(CategoricalAttribute)
				break
			default:
				return nil, fmt.Errorf("Unrecognised Attribute format: %s", a.Type)
			}
			err = attr.UnmarshalJSON(a.Attr)
			if err != nil {
				return nil, fmt.Errorf("Can't deserialize: %s (error: %s)", a, err)
			}
		}
	*/
	return nil, nil
}

func SerializeInstances(inst FixedDataGrid, f io.Writer) error {
	var hdr *tar.Header

	gzWriter := gzip.NewWriter(f)
	tw := tar.NewWriter(gzWriter)

	// Write the MANIFEST entry
	hdr = &tar.Header{
		Name: "MANIFEST",
		Size: int64(len(SerializationFormatVersion)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("Could not write MANIFEST header: %s", err)
	}

	if _, err := tw.Write([]byte(SerializationFormatVersion)); err != nil {
		return fmt.Errorf("Could not write MANIFEST contents: %s", err)
	}

	// Now write the dimensions of the dataset
	attrCount, rowCount := inst.Size()
	hdr = &tar.Header{
		Name: "DIMS",
		Size: 16,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("Could not write DIMS header: %s", err)
	}

	if _, err := tw.Write(PackU64ToBytes(uint64(attrCount))); err != nil {
		return fmt.Errorf("Could not write DIMS (attrCount): %s", err)
	}
	if _, err := tw.Write(PackU64ToBytes(uint64(rowCount))); err != nil {
		return fmt.Errorf("Could not write DIMS (rowCount): %s", err)
	}

	// Write the ATTRIBUTES files
	classAttrs := inst.AllClassAttributes()
	normalAttrs := NonClassAttributes(inst)
	if err := writeAttributesToFilePart(classAttrs, tw, "CATTRS"); err != nil {
		return fmt.Errorf("Could not write CATTRS: %s", err)
	}
	if err := writeAttributesToFilePart(normalAttrs, tw, "ATTRS"); err != nil {
		return fmt.Errorf("Could not write ATTRS: %s", err)
	}

	// Data must be written out in the same order as the Attributes
	allAttrs := make([]Attribute, attrCount)
	normCount := copy(allAttrs, normalAttrs)
	for i, v := range classAttrs {
		allAttrs[normCount+i] = v
	}

	allSpecs := ResolveAttributes(inst, allAttrs)

	// First, estimate the amount of data we'll need...
	dataLength := int64(0)
	inst.MapOverRows(allSpecs, func(val [][]byte, row int) (bool, error) {
		for _, v := range val {
			dataLength += int64(len(v))
		}
		return true, nil
	})

	// Then write the header
	hdr = &tar.Header{
		Name: "DATA",
		Size: dataLength,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("Could not write DATA: %s", err)
	}

	// Then write the actual data
	writtenLength := int64(0)
	if err := inst.MapOverRows(allSpecs, func(val [][]byte, row int) (bool, error) {
		for _, v := range val {
			wl, err := tw.Write(v)
			writtenLength += int64(wl)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}); err != nil {
		return err
	}

	if writtenLength != dataLength {
		return fmt.Errorf("Could not write DATA: changed size from %v to %v", dataLength, writtenLength)
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("Could not flush tar: %s", err)
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("Could not close tar: %s", err)
	}

	if err := gzWriter.Flush(); err != nil {
		return fmt.Errorf("Could not flush gz: %s", err)
	}

	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("Could not close gz: %s", err)
	}

	return nil
}
