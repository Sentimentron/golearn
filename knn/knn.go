// Package KNN implements a K Nearest Neighbors object, capable of both classification
// and regression. It accepts data in the form of a slice of float64s, which are then reshaped
// into a X by Y matrix.
package knn

import (
	"fmt"
	base "github.com/sjwhitworth/golearn/base"
	"math"
	"sort"
)

type neighbour struct {
	rowNumber int
	distance float64
}

type neighbourList []neighbour
func (n neighbourList) Len() int {
	return len(n)
}
func (n neighbourList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
func (n neighbourList) Less(i, j int) bool {
	return n[i].distance > n[j].distance
}

// A KNN Classifier. Consists of a data matrix, associated labels in the same order as the matrix, and a distance function.
// The accepted distance functions at this time are 'euclidean' and 'manhattan'.
type KNNClassifier struct {
	base.BaseEstimator
	TrainingData      *base.Instances
	DistanceFunc      string
	NearestNeighbours int
}

// Returns a new classifier
func NewKnnClassifier(distfunc string, neighbours int) *KNNClassifier {
	KNN := KNNClassifier{}
	KNN.DistanceFunc = distfunc
	KNN.NearestNeighbours = neighbours
	return &KNN
}

// Train stores the training data for llater
func (KNN *KNNClassifier) Fit(trainingData *base.Instances) {
	KNN.TrainingData = trainingData
}

func (KNN *KNNClassifier) Predict(what *base.Instances) base.UpdatableDataGrid {
	var classAttr base.Attribute
	// Generate the prediction vector
	ret := base.GeneratePredictionVector(what)

	// Process the attributes
	classAttrs := what.GetClassAttrs()
	normalAttrs := what.GetAttrs()
	allAttrs := what.GetAttrs()

	// Weed out all the classes
	for attr := range classAttrs {
		classAttr = classAttrs[attr]
		delete(normalAttrs, attr)
	}
	// Weed out all the non-FloatAttributes
	for attr := range allAttrs {
		if _, ok := allAttrs[attr].(*base.FloatAttribute); !ok {
			delete(normalAttrs, attr)
		}
	}

	// Map over the rows
	neighbours := neighbourList(make([]neighbour, KNN.NearestNeighbours+1))
	maxMap := make(map[string] int)
	what.MapOverRows(normalAttrs, func(pred [][]byte, predRow int) (bool, error) {
		for i := 0; i < 3; i++ {
			neighbours[i].distance = math.Inf(1)
		}
		for a := range maxMap {
			maxMap[a] = 0
		}
		// For each item in training...
		KNN.TrainingData.MapOverRows(normalAttrs, func(train [][]byte, trainRow int) (bool, error) {
			distance := 0.0
			for a := range train {
				thisVal := base.UnpackBytesToFloat(train[a])
				otherVal := base.UnpackBytesToFloat(pred[a])
				distance += (thisVal - otherVal) * (thisVal - otherVal)
			}
			neighbours[KNN.NearestNeighbours] = neighbour{trainRow, distance}
			sort.Sort(neighbours)
			return true, nil
		})

		for i := 0; i < KNN.NearestNeighbours; i++ {
			rowNumber := neighbours[i].rowNumber
			label, _ := base.GetClass(KNN.TrainingData, rowNumber)
			maxMap[label]++
		}

		maxClass := ""
		maxVal := 0
		for i := range maxMap {
			if maxMap[i] > maxVal {
				maxClass = i
				maxVal = maxMap[i]
			}
		}

		ret.AppendRowExplicit(map[base.Attribute][]byte{classAttr: classAttr.GetSysValFromString(maxClass)})
		fmt.Println(predRow)
		if predRow >= 20 {
			return false, nil
		}
		return true, nil
	})

	return ret
}

/*
//A KNN Regressor. Consists of a data matrix, associated result variables in the same order as the matrix, and a name.
type KNNRegressor struct {
	base.BaseEstimator
	Values       []float64
	DistanceFunc string
}

// Mints a new classifier.
func NewKnnRegressor(distfunc string) *KNNRegressor {
	KNN := KNNRegressor{}
	KNN.DistanceFunc = distfunc
	return &KNN
}

func (KNN *KNNRegressor) Fit(values []float64, numbers []float64, rows int, cols int) {
	if rows != len(values) {
		panic(mat64.ErrShape)
	}

	KNN.Data = mat64.NewDense(rows, cols, numbers)
	KNN.Values = values
}

func (KNN *KNNRegressor) Predict(vector *mat64.Dense, K int) float64 {

	// Get the number of rows
	rows, _ := KNN.Data.Dims()
	rownumbers := make(map[int]float64)
	labels := make([]float64, 0)

	// Check what distance function we are using
	switch KNN.DistanceFunc {
	case "euclidean":
		{
			euclidean := pairwiseMetrics.NewEuclidean()
			for i := 0; i < rows; i++ {
				row := KNN.Data.RowView(i)
				rowMat := util.FloatsToMatrix(row)
				distance := euclidean.Distance(rowMat, vector)
				rownumbers[i] = distance
			}
		}
	case "manhattan":
		{
			manhattan := pairwiseMetrics.NewEuclidean()
			for i := 0; i < rows; i++ {
				row := KNN.Data.RowView(i)
				rowMat := util.FloatsToMatrix(row)
				distance := manhattan.Distance(rowMat, vector)
				rownumbers[i] = distance
			}
		}
	}

	sorted := util.SortIntMap(rownumbers)
	values := sorted[:K]

	var sum float64
	for _, elem := range values {
		value := KNN.Values[elem]
		labels = append(labels, value)
		sum += value
	}

	average := sum / float64(K)
	return average
}*/
