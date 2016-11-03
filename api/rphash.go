package api

import (
	"container/list"
	"encoding/gob"
	"flag"
	"sync"

	"github.com/chrislusf/glow/flow"
	"github.com/wilseypa/rphash-golang/itemset"
	"github.com/wilseypa/rphash-golang/parse"
	"github.com/wilseypa/rphash-golang/reader"
	"github.com/wilseypa/rphash-golang/stream"
	"github.com/wilseypa/rphash-golang/utils"
)

type CentroidData struct {
	dimensions int
	matrixList *list.List
}

// Function used to combine the data in the list (of the CentroidData structure)
// Into a single matrix (type: [][]float64).
func (this *CentroidData) getUnderlyingMatrix() [][]float64 {

	// Determine the final dimensions of the matrixList
	totalSize := 0
	for e := this.matrixList.Front(); e != nil; e = e.Next() {
		totalSize += len(e.Value.([][]float64))
	}

	// Allocate the new matrix
	newMatrix := make([][]float64, totalSize)
	//for i := range newMatrix {
	//	newMatrix[i] = make([]float64, this.dimensions)
	//}

	// Transfer over the list items to the matrix
	indx := 0
	for e := this.matrixList.Front(); e != nil; e = e.Next() {
		tmpMatrix := e.Value.([][]float64)
		for i := 0; i < len(tmpMatrix); i++ {
			newMatrix[indx] = tmpMatrix[i]
			indx++
		}
	}

	// Return the new marix
	return newMatrix
}

type Centroid struct {
	C *itemset.Centroid
}

func goStart(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

func ClusterFile(filename string, distributed bool, clusters int) [][]float64 {
	data := utils.ReadCSV(filename)
	if distributed {
		return ClusterDist(data, clusters)
	} else {
		return Cluster(data)
	}
}

func ClusterDist(records [][]float64, clusters int) [][]float64 {

	// Create the structure to hold the results
	initList := list.New()
	centroidData := CentroidData{0, initList}

	// Create a new Go-Flow mapping for processing the data in distributed form.
	flow.New().Source(func(out chan [][]float64) {

		// Assign the data chunks to channels.
		size := len(records)
		unit := int(size / clusters)
		for i := 0; i < clusters; i++ {
			start := i * unit
			end := (i + 1) * unit
			if end >= size {
				end = size - 1
			}
			currSlice := records[start:end][:]
			out <- currSlice
		}

	}, clusters).Map(func(dataSlice [][]float64) [][]float64 {
		return Cluster(dataSlice)
	}).Map(func(dataSlice [][]float64) {
		centroidData.matrixList.PushBack(dataSlice)
	}).Run()

	// Return the centroid results
	return centroidData.getUnderlyingMatrix()
}

func Cluster(records [][]float64) [][]float64 {

	f := flow.New()
	numClusters := 6

	gob.Register(Centroid{})
	gob.Register(itemset.Centroid{})
	gob.Register(utils.Hash64Set{})
	flag.Parse()

	Object := reader.NewStreamObject(len(records[0]), numClusters)
	Stream := stream.NewStream(Object)

	outChannel := make(chan Centroid)

	ch := make(chan []float64)

	source := f.Channel(ch)

	f1 := source.Map(func(record []float64) Centroid {
		return Centroid{C: Stream.AddVectorOnlineStep(record)}
	}).AddOutput(outChannel)

	flow.Ready()

	var wg sync.WaitGroup

	goStart(&wg, func() {
		f1.Run()
	})

	goStart(&wg, func() {
		for out := range outChannel {
			Stream.CentroidCounter.Add(out.C)
		}
	})

	for _, record := range records {
		ch <- record
	}

	close(ch)
	wg.Wait()

	return Stream.GetCentroids()
}

func Denormalize(dimension float64) float64 {
	return parse.DeNormalize(dimension)
}
