package main

import (
	"reflect"
	"testing"
)

func TestDFSIterativeStack(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(1, 4)

	result := graph.DFSIterativeStack(1)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) && !reflect.DeepEqual(result, []int{1, 4, 2, 3}) {
		t.Errorf("DFSIterativeStack got %v, want %v", result, expected)
	}
}

func TestDFSRecursive(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(1, 4)

	result := graph.DFSRecursive(1)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result, expected) && !reflect.DeepEqual(result, []int{1, 4, 2, 3}) {
		t.Errorf("DFSRecursive got %v, want %v", result, expected)
	}
}

func TestBFSQueue(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(1, 4)

	result := graph.BFSQueue(1)
	expected := []int{1, 2, 4, 3}
	if !reflect.DeepEqual(result, expected) && !reflect.DeepEqual(result, []int{1, 4, 2, 3}) {
		t.Errorf("BFSQueue got %v, want %v", result, expected)
	}
}

func TestDFSFindPath(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(1, 4)

	path := graph.DFSFindPath(1, 3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(path, expected) {
		t.Errorf("DFSFindPath got %v, want %v", path, expected)
	}
}

func TestBFSShortestPath(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(1, 4)
	graph.AddEdge(4, 3)

	path := graph.BFSShortestPath(1, 3)
	expected := []int{1, 4, 3}
	if !reflect.DeepEqual(path, expected) && !reflect.DeepEqual(path, []int{1, 2, 3}) {
		t.Errorf("BFSShortestPath got %v, want %v", path, expected)
	}
}

func TestDetectCycleDFS(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(3, 1)
	if !graph.DetectCycleDFS() {
		t.Errorf("DetectCycleDFS should detect a cycle")
	}

	graph2 := NewGraph()
	graph2.AddEdge(1, 2)
	graph2.AddEdge(2, 3)
	if graph2.DetectCycleDFS() {
		t.Errorf("DetectCycleDFS should not detect a cycle")
	}
}

func TestTopologicalSortDFS(t *testing.T) {
	dag := NewGraph()
	dag.AddDirectedEdge(1, 2)
	dag.AddDirectedEdge(1, 3)
	dag.AddDirectedEdge(2, 4)
	dag.AddDirectedEdge(3, 4)
	dag.AddDirectedEdge(4, 5)

	order := dag.TopologicalSortDFS()
	// Valid topological orders: [1 3 2 4 5], [1 2 3 4 5], etc.
	// Just check that 1 comes before 2 and 3, 2 and 3 before 4, 4 before 5
	pos := map[int]int{}
	for i, v := range order {
		pos[v] = i
	}
	if !(pos[1] < pos[2] && pos[1] < pos[3] && pos[2] < pos[4] && pos[3] < pos[4] && pos[4] < pos[5]) {
		t.Errorf("TopologicalSortDFS order invalid: %v", order)
	}
}

func TestConnectedComponentsDFS(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge(1, 2)
	graph.AddEdge(3, 4)
	components := graph.ConnectedComponentsDFS()
	if len(components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(components))
	}
}
