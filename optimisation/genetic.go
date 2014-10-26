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
	pool := make([]survivalRecord, top)

	// Compute initial fitness
	f := fitness(template)

	// Populate the survival pool
	for i := range pool {
		pool[i].g = template.Copy()
		pool[i].f = f
	}

	// The first go-routine continually selects and breeds pairs
	// from the pool.
	n := runtime.NumCPU()
	stork := make(chan Genome, n)
	fittest := make(chan survivalRecord, n*32)
	done := 0

	// Evaluation is what normally takes up the time
	var processWait sync.WaitGroup
	for i := 0; i < n; i++ {
		processWait.Add(1)
		go func() {
			more := true
			for {
				select {
				case cur := <-stork:
					if cur == nil {
						more = false
						break
					}
					f := fitness(cur)
					fittest <- survivalRecord{cur, f}
				default:
					more = false
					break
				}
				if done == rounds {
					processWait.Done()
					break
				}
				if !more {
					break
				}
			}
		}()
	}

	func() {
		for r := 0; r < rounds; r++ {
			done++
			// Process any new fitness records
			for {
				more := true
				select {
				case i := <-fittest:
					minFitness := math.Inf(1)
					minIndex := 0
					for i := range pool {
						if pool[i].f < minFitness {
							minIndex = i
							minFitness = pool[i].f
						}
					}
					f := i.f
					if f > minFitness {
						pool[minIndex] = i
					}
				default:
					more = false
				}
				if !more {
					break
				}
			}

			// Select two random genomes
			i := 0
			j := 0
			for {
				if i != j {
					break
				}
				i = rand.Intn(top)
				j = rand.Intn(top)
			}
			fmt.Println("Breeding")
			g := pool[i].g.Breed(pool[j].g)
			g.Randomize(alpha)
			stork <- g
		}
		close(stork)
	}()

	processWait.Wait()

	fmt.Println(pool)
	// Process any new fitness records
	for {
		more := true
		select {
		case i := <-fittest:
			minFitness := math.Inf(1)
			minIndex := 0
			for i := range pool {
				if pool[i].f < minFitness {
					minIndex = i
					minFitness = pool[i].f
				}
			}
			f := i.f
			if f > minFitness {
				pool[minIndex] = i
			}
		default:
			more = false
		}
		if !more {
			break
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
