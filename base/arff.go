package base

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// GetARFFDataRowCount returns the number of data rows in an ARFF file.
func GetARFFDataRowCount(filepath string) int {

	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	counting := false
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if counting {
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
	return count
}

// GetARFFAttributes returns the set of Attributes represented in this ARFF
func GetARFFAttributes(filepath string) []Attribute {
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
		case "numeric":
			attr = new(FloatAttribute)
			break
		case "binary":
			attr = new(BinaryAttribute)
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
					attr := NewCategoricalAttribute()
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
		attr.SetName(fields[1])
		ret = append(ret, attr)
	}

	return ret
}

// ParseARFFToTemplatedInstances parses the dense ARFF file into a FixedDataGrid, using only the given parameters.
func ParseARFFToTemplatedInstances(filepath string, template *DenseInstances) (instances *DenseInstances, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	attrs := GetARFFAttributes(filepath)

	dataSection := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ToLower(line)
		line = strings.TrimSpace(line)
		if line == "@data" {
			dataSection = true
			break
		}
	}

	if !dataSection {
		return nil, fmt.Errorf("No @data section!")
	}

	instances = CopyDenseInstances(template, attrs)
	err = ParseCSVBuildInstancesFromReader(f, false, instances)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

// ParseARFFToInstances parses the dense ARFF File into a FixedDataGrid
func ParseARFFToInstances(filepath string) (instances *DenseInstances, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	// Find the number of rows in the file
	rows := GetARFFDataRowCount(filepath)

	// Get the Attributes we want
	attrs := GetARFFAttributes(filepath)

	// Allocate return value
	ret := NewDenseInstances()
	ret.Extend(rows)

	// Add all the Attributes
	for _, a := range attrs {
		ret.AddAttribute(a)
	}

	// Set the last Attribute as the class
	ret.AddClassAttribute(attrs[len(attrs)-1])

	// Read the data
	return ParseARFFToTemplatedInstances(filepath, instances)
}
