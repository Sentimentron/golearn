package optimisation

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

type empty struct{}

// Genome specifies what every living organism must do.
type Genome interface {
	Copy() Genome
	Randomize(float64)
	Breed(Genome) Genome
}

// BasicGenome is a vector of floats, maps nicely onto randomization.
type BasicGenome struct {
	Vals []float64
}

// Copy returns a BasicGenome identical to its parent.
func (b *BasicGenome) Copy() Genome {
	ret := make([]float64, len(b.Vals))
	copy(ret, b.Vals)
	return &BasicGenome{
		ret,
	}
}

// Randomize randomizes some values in place.
func (b *BasicGenome) Randomize(variance float64) {
	for i := range b.Vals {
		b.Vals[i] += rand.NormFloat64() * variance
	}
}

// Breed crosses this Genome with another.
func (b *BasicGenome) Breed(other Genome) Genome {
	var a *BasicGenome
	var ok bool
	if a, ok = other.(*BasicGenome); !ok {
		panic("Incompatable organisms")
	}
	newGenome := b.Copy().(*BasicGenome)
	for i := range b.Vals {
		if rand.Intn(2) == 1 {
			newGenome.Vals[i] = a.Vals[i]
		}
	}
	return newGenome
}

type survivalRecord struct {
	g Genome
	f float64
}

// BasicGenomeOptimize creates a pool of n randomized initial Genomes
// copied from the template. It then repeatedly allows these guys to
// breed, evaluating their fitneess each time.
func BasicGenomeOptimize(template Genome, top, rounds int, fitness func(Genome) float64, alpha float64) Genome {

	// Create the initial pool
	pool := make([]*survivalRecord, top)

	// Compute initial fitness
	f := fitness(template)

	// Populate the survival pool
	for i := range pool {
		pool[i] = &survivalRecord{nil, 0.0}
		pool[i].g = template.Copy()
		pool[i].f = f
	}

	// The first go-routine continually selects and breeds pairs
	// from the pool.
	n := runtime.NumCPU()
	stork := make([]*survivalRecord, n)

	for i := 0; i < rounds; i += n {
		var processWait sync.WaitGroup
		// Breed each thing
		for l := range stork {
			// Choose two parent genomes
			j := 0
			k := 0
			for {
				if j == k {
					break
				}
				j = rand.Intn(top)
				k = rand.Intn(top)
			}
			stork[l] = &survivalRecord{nil, 0.0}
			// Breed
			stork[l].g = pool[j].g.Breed(pool[k].g)
			stork[l].g.Randomize(alpha)
		}
		// Evaluate fitness
		for j := 0; j < n; j++ {
			processWait.Add(1)
			go func(k int, cur *survivalRecord) {
				fmt.Println(k, cur)
				f := fitness(cur.g)
				stork[k].f = f
				processWait.Done()
			}(j, stork[j])
		}
		processWait.Wait()

		// Update the pool
		for _, g := range stork {
			minFitness := math.Inf(1)
			minIndex := 0
			for j, p := range pool {
				if p.f < minFitness {
					minFitness = p.f
					minIndex = j
				}
			}
			if minFitness < g.f {
				pool[minIndex] = g
			}
		}
	}

	maxFitness := math.Inf(-1)
	maxIndex := 0
	for i := range pool {
		fmt.Println(pool[i].f, maxFitness)
		if pool[i].f > maxFitness {
			maxIndex = i
			maxFitness = pool[i].f
		}
	}

	return pool[maxIndex].g

}
