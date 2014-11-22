package evaluation

import (
	"github.com/sjwhitworth/golearn/base"
	"math/rand"
	"sync"
)

// GetCrossValidatedMetric returns the mean and variance of the confusion-matrix-derived
// metric across all folds.
func GetCrossValidatedMetric(in []ConfusionMatrix, metric func(ConfusionMatrix) float64) (mean, variance float64) {
	scores := make([]float64, len(in))
	for i, c := range in {
		scores[i] = metric(c)
	}

	// Compute mean, variance
	sum := 0.0
	for _, s := range scores {
		sum += s
	}
	sum /= float64(len(scores))
	mean = sum
	sum = 0.0
	for _, s := range scores {
		sum += (s - mean) * (s - mean)
	}
	sum /= float64(len(scores))
	variance = sum
	return mean, variance
}

// GenerateCrossFoldValidationConfusionMatrices divides the data into a number of folds
// then trains and evaluates the classifier on each fold, producing a new ConfusionMatrix.
func GenerateCrossFoldValidationConfusionMatrices(data base.FixedDataGrid, cls base.Classifier, folds int) ([]ConfusionMatrix, error) {
	_, rows := data.Size()

	// Assign each row to a fold
	foldMap := make([]int, rows)
	inverseFoldMap := make(map[int][]int)
	for i := 0; i < rows; i++ {
		fold := rand.Intn(folds)
		foldMap[i] = fold
		if _, ok := inverseFoldMap[fold]; !ok {
			inverseFoldMap[fold] = make([]int, 0)
		}
		inverseFoldMap[fold] = append(inverseFoldMap[fold], i)
	}

	ret := make([]ConfusionMatrix, folds)

	// Create training/test views for each fold
	var wait sync.WaitGroup
	for i := 0; i < folds; i++ {
		wait.Add(1)
		go func(i int) {
			// Fold i is for testing
			testData := base.SelectRowsFromMap(data, inverseFoldMap[i])
			otherRows := make([]int, 0)
			for j := 0; j < folds; j++ {
				if i == j {
					continue
				}
				otherRows = append(otherRows, inverseFoldMap[j]...)
			}

			trainData := base.SelectRowsFomMap(data, otherRows)
			// Train
			clsFld := cls.CopyUntrained()
			err := clsFld.Fit(trainData)
			if err != nil {
				panic(err)
			}
			// Predict
			pred, err := clsFld.Predict(testData)
			if err != nil {
				panic(err)
			}
			// Evaluate
			cf, err := GetConfusionMatrix(testData, pred)
			if err != nil {
				panic(err)
			}
			ret[i] = cf
			wait.Done()
		}(i)
	}
	wait.Wait()

	return ret, nil

}
