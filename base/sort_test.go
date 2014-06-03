package base

import "fmt"
import "testing"

func isSortedAsc(inst *Instances, attrIndex int) bool {
	valPrev := 0.0
	for i := 0; i < inst.Rows; i++ {
		cur := inst.get(i, attrIndex)
		if i > 0 {
			if valPrev > cur {
				return false
			}
		}
		valPrev = cur
	}
	return true
}

func isSortedDesc(inst *Instances, attrIndex int) bool {
	valPrev := 0.0
	for i := 0; i < inst.Rows; i++ {
		cur := inst.get(i, attrIndex)
		if i > 0 {
			if valPrev < cur {
				return false
			}
		}
		valPrev = cur
	}
	return true
}

func TestAttrLookup(testEnv *testing.T) {
	inst1, err := ParseCSVToInstances("../examples/datasets/iris_headers.csv", true)
	if err != nil {
		testEnv.Error(err)
		return
	}
	attributes := inst1.GetAttrs()
	attrs := make([]Attribute, 4)
	attrs[0] = attributes[3]
	attrs[1] = attributes[2]
	attrs[2] = attributes[1]
	attrs[3] = attributes[0]
	resolvedAttrs, err := inst1.resolveToInternal(attrs)
	if err != nil {
		testEnv.Error(err)
	}
	for i, j := range resolvedAttrs {
		if i != 3-j {
			testEnv.Error(fmt.Sprintf("%d %d", i, j))
		}
	}
}

func TestSortDesc(testEnv *testing.T) {
	inst1, err := ParseCSVToInstances("../examples/datasets/iris_headers.csv", true)
	if err != nil {
		testEnv.Error(err)
		return
	}
	inst2, err := ParseCSVToInstances("../examples/datasets/iris_sorted_desc.csv", true)
	if err != nil {
		testEnv.Error(err)
		return
	}

	if isSortedDesc(inst1, 0) {
		testEnv.Error("Can't test descending sort order")
	}
	if !isSortedDesc(inst2, 0) {
		testEnv.Error("Reference data not sorted in descending order!")
	}

	attributes := inst1.GetAttrs()
	attrs := make([]Attribute, 4)
	attrs[0] = attributes[3]
	attrs[1] = attributes[2]
	attrs[2] = attributes[1]
	attrs[3] = attributes[0]
	err = inst1.Sort(Descending, attrs)
	if err != nil {
		testEnv.Error(err)
		return
	}
	if !isSortedDesc(inst1, 0) {
		testEnv.Error("Instances are not sorted in descending order")
		testEnv.Error(inst1)
	}
	if !inst2.Equals(inst1) {
		inst1.storage.Sub(inst1.storage, inst2.storage)
		testEnv.Error("Instances don't match")
		testEnv.Error(inst1)
		testEnv.Error(inst2)
	}
}

func TestSortAsc(testEnv *testing.T) {
	inst, err := ParseCSVToInstances("../examples/datasets/iris_headers.csv", true)
	if isSortedAsc(inst, 0) {
		testEnv.Error("Can't test ascending sort on something ascending already")
	}
	if err != nil {
		testEnv.Error(err)
		return
	}
	attributes := inst.GetAttrs()
	attrs := make([]Attribute, 4)
	attrs[0] = attributes[3]
	attrs[1] = attributes[2]
	attrs[2] = attributes[1]
	attrs[3] = attributes[0]
	inst.Sort(Ascending, attrs)
	if !isSortedAsc(inst, 0) {
		testEnv.Error("Instances are not sorted in ascending order")
		testEnv.Error(inst)
	}

	inst2, err := ParseCSVToInstances("../examples/datasets/iris_sorted_asc.csv", true)
	if err != nil {
		testEnv.Error(err)
		return
	}
	if !isSortedAsc(inst2, 0) {
		testEnv.Error("This file should be sorted in ascending order")
	}

	if !inst2.Equals(inst) {
		inst.storage.Sub(inst.storage, inst2.storage)
		testEnv.Error(inst.storage)
		testEnv.Error("Instances don't match")
		testEnv.Error(inst)
		testEnv.Error(inst2)
	}

}
