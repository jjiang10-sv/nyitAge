package graph

import (
	"fmt"
	"math"
)

type EdgeBell struct {
	source, destination, weight int
}

// BellmanFord finds shortest paths from src to all vertices.
// Returns an error if a negative weight cycle is detected.
func BellmanFord(vertices int, edges []EdgeBell, src int) ([]int, error) {
	// Initialize distances from src to all other vertices as infinity
	distances := make([]int, vertices)
	for i := range distances {
		distances[i] = math.MaxInt32
	}
	distances[src] = 0

	// Relax edges |V| - 1 times
	for i := 0; i < vertices-1; i++ {
		for _, edge := range edges {
			if distances[edge.source] != math.MaxInt32 && distances[edge.source]+edge.weight < distances[edge.destination] {
				distances[edge.destination] = distances[edge.source] + edge.weight
			}
		}
	}

	// Check for negative-weight cycles
	for _, edge := range edges {
		if distances[edge.source] != math.MaxInt32 && distances[edge.source]+edge.weight < distances[edge.destination] {
			fmt.Println("affected nodes", edge.source, edge.destination)
			return nil, fmt.Errorf("graph contains a negative weight cycle")
		}
	}

	return distances, nil
}

type BellmanFord1001 struct {
	edges []EdgeBell
	vertices int
}
func (b *BellmanFord1001) findShortestPath(src int) ([]int, error) {
	result := []int{}
	for i:=0; i < b.vertices;i++{
		result[i] = math.MaxInt
	}
	result[src] = 0
	for i := 0; i < b.vertices-1;i++ {
		for _, edge := range b.edges {
			if result[edge.destination] > result[edge.source] + edge.weight && result[edge.source] != math.MaxInt{
				result[edge.destination] = result[edge.source] + edge.weight
			}
		}
	}
	for _, edge := range b.edges{
		if result[edge.source] != math.MaxInt && result[edge.source]+edge.weight < result[edge.destination]{
			return nil, fmt.Errorf("there is a negtive weight in the graph")
		}
	}
	return result, nil
}


func main() {
	vertices := 5
	edges := []EdgeBell{
		{0, 1, -1},
		{0, 2, 4},
		{1, 2, 3},
		{1, 3, 2},
		{1, 4, 2},
		{3, 2, 5},
		{3, 1, 1},
		{4, 3, -3},
	}

	// Run Bellman-Ford from vertex 0
	distances, err := BellmanFord(vertices, edges, 0)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Vertex distances from source:")
		for i, d := range distances {
			fmt.Printf("Distance to vertex %d: %d\n", i, d)
		}
	}
}
