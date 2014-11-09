package base

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

// ParseARFFGetRows returns the number of data rows in an ARFF file.
func ParseARFFGetRows(filepath string) (int, error) {

	f, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	counting := false
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if counting {
			if line[0] == '@' {
				continue
			}
			if line[0] == '%' {
				continue
			}
			count++
			continue
		}
		if line[0] == '@' {
			line = strings.ToLower(line)
			if line == "@data" {
				counting = true
			}
		}
	}
	return count, nil
}

// ParseARFFGetAttributes returns the set of Attributes represented in this ARFF
func ParseARFFGetAttributes(filepath string) []Attribute {
	var ret []Attribute

	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var attr Attribute
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] != '@' {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		fields[0] = strings.ToLower(fields[0])
		fields[2] = strings.ToLower(fields[2])
		if fields[0] != "@attribute" {
			continue
		}
		switch fields[2] {
		case "real":
			attr = new(FloatAttribute)
			break
		default:
			if fields[2][0] == '{' {
				if fields[2][len(fields[2])-1] == '}' {
					cats := strings.Split(fields[2][1:len(fields[2])-1], ",")
					if len(cats) == 0 {
						panic(fmt.Errorf("Empty categorical field on line '%s'", line))
					}
					for i, v := range cats {
						cats[i] = strings.TrimSpace(v)
					}
					attr = NewCategoricalAttribute()
					for _, v := range cats {
						attr.GetSysValFromString(v)
					}
				} else {
					panic(fmt.Errorf("Missing categorical bracket on line '%s'", line))
				}
			} else {
				panic(fmt.Errorf("Unsupported Attribute type %s on line '%s'", fields[2], line))
			}
		}
		if attr == nil {
			panic(fmt.Errorf(line))
		}
		attr.SetName(fields[1])
		ret = append(ret, attr)
	}

	maxPrecision, err := ParseCSVEstimateFilePrecision(filepath)
	if err != nil {
		panic(err)
	}
	for _, a := range ret {
		if f, ok := a.(*FloatAttribute); ok {
			f.Precision = maxPrecision
		}
	}

	return ret
}

func ParseARFFGetHeaderSize(r io.Reader) int64 {
	var read int64
	reader := bufio.NewScanner(r)
	for reader.Scan() {
		line := reader.Text()
		read += int64(len(line))
		if len(line) == 0 {
			continue
		}
		if line[0] == '@' {
			line = strings.ToLower(line)
			line = strings.TrimSpace(line)
			if line == "@data" {
				break
			}
			continue
		} else if line[0] == '%' {
			continue
		}
	}
	return read
}

// ParseDenseARFFBuildInstancesFromReader updates an [[#UpdatableDataGrid]] from a io.Reader
func ParseDenseARFFBuildInstancesFromReader(r io.Reader, u UpdatableDataGrid) (err error) {
	var rowCounter int

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(err)
			}
			err = fmt.Errorf("Error at line %d (error %s)", rowCounter, r.(error))
		}
	}()

	scanner := bufio.NewScanner(r)
	reading := false
	specs := ResolveAttributes(u, u.AllAttributes())
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "%") {
			continue
		}
		if reading {
			buf := bytes.NewBuffer([]byte(line))
			reader := csv.NewReader(buf)
			for {
				r, err := reader.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}
				for i, v := range r {
					u.Set(specs[i], rowCounter, specs[i].attr.GetSysValFromString(v))
				}
				rowCounter++
			}
		} else {
			line = strings.ToLower(line)
			line = strings.TrimSpace(line)
			if line == "@data" {
				reading = true
			}
		}
	}

	return nil
}

// ParseDenseARFFToInstances parses the dense ARFF File into a FixedDataGrid
func ParseDenseARFFToInstances(filepath string) (ret *DenseInstances, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	// Find the number of rows in the file
	rows, err := ParseARFFGetRows(filepath)
	if err != nil {
		return nil, err
	}

	// Get the Attributes we want
	attrs := ParseARFFGetAttributes(filepath)

	// Allocate return value
	ret = NewDenseInstances()

	// Add all the Attributes
	for _, a := range attrs {
		ret.AddAttribute(a)
	}

	// Set the last Attribute as the class
	ret.AddClassAttribute(attrs[len(attrs)-1])
	ret.Extend(rows)

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read the data
	// Seek past the header
	err = ParseDenseARFFBuildInstancesFromReader(f, ret)
	if err != nil {
		ret = nil
	}
	return ret, err
}
