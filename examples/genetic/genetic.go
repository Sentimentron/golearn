package main

import (
	"fmt"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/ensemble"
	"github.com/sjwhitworth/golearn/evaluation"
	"github.com/sjwhitworth/golearn/optimisation"
)

func main() {
	inst, err := base.ParseCSVToInstances("../datasets/articles.csv", false)
	if err != nil {
		panic(err)
	}

	X, Y := base.InstancesTrainTestSplit(inst, 0.4)

	fitness := func(g optimisation.Genome) float64 {
		b := g.(*optimisation.BasicGenome)
		weights := make(map[string]float64)
		for i := range b.Vals {
			if b.Vals[i] < 0.1 {
				b.Vals[i] = 0.1
			}
		}
		weights["Finance"] = b.Vals[0]
		weights["Politics"] = b.Vals[1]
		weights["Tech"] = b.Vals[2]
		errorTerm := b.Vals[3]
		m := ensemble.NewMultiLinearSVC("l1", "l2", true, errorTerm, 1e-4, weights)
		m.Fit(X)
		predictions, err := m.Predict(Y)
		if err != nil {
			panic(err)
		}
		cf, err := evaluation.GetConfusionMatrix(Y, predictions)
		if err != nil {
			panic(err)
		}
		f := evaluation.GetAccuracy(cf)
		fmt.Println(weights, f)
		return f
	}

	initialGenome := new(optimisation.BasicGenome)
	//	initialGenome.Vals = []float64{1.0, 1.0, 1.0, 1.0}
	initialGenome.Vals = []float64{0.94, 0.98, 0.81, 0.81}

	optimizedGenome := optimisation.BasicGenomeOptimize(initialGenome, 15, 200, fitness, 0.105)
	finalFitness := fitness(optimizedGenome)
	fmt.Println(finalFitness)
	fmt.Println(optimizedGenome.(*optimisation.BasicGenome).Vals)
}
