package main

import (
	"container/list"
	"fmt"
)

// Graph represents an undirected graph using adjacency list
type Graph struct {
	adjacencyList map[int][]int
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		adjacencyList: make(map[int][]int),
	}
}

// AddEdge adds an undirected edge between nodes u and v
func (g *Graph) AddEdge(u, v int) {
	g.adjacencyList[u] = append(g.adjacencyList[u], v)
	g.adjacencyList[v] = append(g.adjacencyList[v], u)
}

// AddDirectedEdge adds a directed edge from u to v
func (g *Graph) AddDirectedEdge(u, v int) {
	g.adjacencyList[u] = append(g.adjacencyList[u], v)
}

// ============================================================================
// DEPTH-FIRST SEARCH (DFS) WITH STACK
// ============================================================================

// DFSIterativeStack performs DFS using explicit stack
// Time Complexity: O(V + E)
// Space Complexity: O(V)
func (g *Graph) DFSIterativeStack(start int) []int {
	if len(g.adjacencyList) == 0 {
		return []int{}
	}

	visited := make(map[int]bool)
	stack := list.New()
	result := []int{}

	// Push start node onto stack
	stack.PushBack(start)

	for stack.Len() > 0 {
		// Pop from top of stack (LIFO)
		current := stack.Remove(stack.Back()).(int)

		if !visited[current] {
			visited[current] = true
			result = append(result, current)

			// Push all unvisited neighbors onto stack
			// Note: We push in reverse order to maintain natural traversal order
			neighbors := g.adjacencyList[current]
			for i := len(neighbors) - 1; i >= 0; i-- {
				neighbor := neighbors[i]
				if !visited[neighbor] {
					stack.PushBack(neighbor)
				}
			}
		}
	}

	return result
}
func (g *Graph) DFSIterativeStack1(start int) []int {
	result := []int{}
	stack := list.New()
	stack.PushBack(start)
	visited := make(map[int]bool)

	for stack.Len() > 0 {
		current := stack.Remove(stack.Back()).(int)
		if !visited[current] {
			result = append(result, current)
			for _, node := range g.adjacencyList[current] {
				if !visited[node] {
					stack.PushBack(node)
				}
			}
			visited[current] = true
		}

	}
	return result
}

// DFSRecursive performs DFS using recursion (implicit call stack)
func (g *Graph) DFSRecursive(start int) []int {
	visited := make(map[int]bool)
	result := []int{}

	var dfsHelper func(node int)
	dfsHelper = func(node int) {
		if visited[node] {
			return
		}

		visited[node] = true
		result = append(result, node)

		for _, neighbor := range g.adjacencyList[node] {
			if !visited[neighbor] {
				dfsHelper(neighbor)
			}
		}
	}

	dfsHelper(start)
	return result
}

// ============================================================================
// BREADTH-FIRST SEARCH (BFS) WITH QUEUE
// ============================================================================

// BFSQueue performs BFS using explicit queue
// Time Complexity: O(V + E)
// Space Complexity: O(V)
func (g *Graph) BFSQueue(start int) []int {
	if len(g.adjacencyList) == 0 {
		return []int{}
	}

	visited := make(map[int]bool)
	queue := list.New()
	result := []int{}

	// Enqueue start node
	queue.PushBack(start)

	for queue.Len() > 0 {
		// Dequeue from front (FIFO)
		current := queue.Remove(queue.Front()).(int)

		if !visited[current] {
			visited[current] = true
			result = append(result, current)

			// Enqueue all unvisited neighbors
			for _, neighbor := range g.adjacencyList[current] {
				if !visited[neighbor] {
					queue.PushBack(neighbor)
				}
			}
		}
	}

	return result
}

// ============================================================================
// ADVANCED EXAMPLES WITH PATH FINDING
// ============================================================================

// PathNode represents a node with its path from start
type PathNode struct {
	node int
	path []int
}

// DFSFindPath finds a path from start to end using DFS
func (g *Graph) DFSFindPath(start, end int) []int {
	if start == end {
		return []int{start}
	}

	visited := make(map[int]bool)
	stack := list.New()

	// Push start node with its path
	stack.PushBack(PathNode{node: start, path: []int{start}})

	for stack.Len() > 0 {
		current := stack.Remove(stack.Back()).(PathNode)

		if current.node == end {
			return current.path
		}

		if !visited[current.node] {
			visited[current.node] = true

			for _, neighbor := range g.adjacencyList[current.node] {
				if !visited[neighbor] {
					newPath := make([]int, len(current.path))
					copy(newPath, current.path)
					newPath = append(newPath, neighbor)
					stack.PushBack(PathNode{node: neighbor, path: newPath})
				}
			}
		}
	}

	return nil // No path found
}

// BFSShortestPath finds shortest path from start to end using BFS
func (g *Graph) BFSShortestPath(start, end int) []int {
	if start == end {
		return []int{start}
	}

	visited := make(map[int]bool)
	queue := list.New()

	// Enqueue start node with its path
	queue.PushBack(PathNode{node: start, path: []int{start}})

	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(PathNode)

		if current.node == end {
			return current.path
		}

		if !visited[current.node] {
			visited[current.node] = true

			for _, neighbor := range g.adjacencyList[current.node] {
				if !visited[neighbor] {
					newPath := make([]int, len(current.path))
					copy(newPath, current.path)
					newPath = append(newPath, neighbor)
					queue.PushBack(PathNode{node: neighbor, path: newPath})
				}
			}
		}
	}

	return nil // No path found
}

// ============================================================================
// USE CASES AND APPLICATIONS
// ============================================================================

// DetectCycleDFS detects cycle in undirected graph using DFS
func (g *Graph) DetectCycleDFS() bool {
	visited := make(map[int]bool)

	var hasCycleDFS func(node, parent int) bool
	hasCycleDFS = func(node, parent int) bool {
		visited[node] = true

		for _, neighbor := range g.adjacencyList[node] {
			if !visited[neighbor] {
				if hasCycleDFS(neighbor, node) {
					return true
				}
			} else if neighbor != parent {
				return true // Back edge found
			}
		}
		return false
	}

	for node := range g.adjacencyList {
		if !visited[node] {
			if hasCycleDFS(node, -1) {
				return true
			}
		}
	}
	return false
}

// TopologicalSortDFS performs topological sort using DFS (for DAGs)
func (g *Graph) TopologicalSortDFS() []int {
	visited := make(map[int]bool)
	tempVisited := make(map[int]bool) // For cycle detection
	result := []int{}

	var dfsTopo func(node int) bool
	dfsTopo = func(node int) bool {
		if tempVisited[node] {
			return false // Cycle detected
		}
		if visited[node] {
			return true
		}

		tempVisited[node] = true

		for _, neighbor := range g.adjacencyList[node] {
			if !dfsTopo(neighbor) {
				return false
			}
		}

		delete(tempVisited, node)
		visited[node] = true
		result = append(result, node)
		return true
	}

	for node := range g.adjacencyList {
		if !visited[node] {
			if !dfsTopo(node) {
				return []int{} // Cycle detected
			}
		}
	}

	// Reverse to get topological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// ConnectedComponentsDFS finds all connected components using DFS
func (g *Graph) ConnectedComponentsDFS() [][]int {
	visited := make(map[int]bool)
	components := [][]int{}

	var dfsComponent func(node int, component *[]int)
	dfsComponent = func(node int, component *[]int) {
		visited[node] = true
		*component = append(*component, node)

		for _, neighbor := range g.adjacencyList[node] {
			if !visited[neighbor] {
				dfsComponent(neighbor, component)
			}
		}
	}

	for node := range g.adjacencyList {
		if !visited[node] {
			component := []int{}
			dfsComponent(node, &component)
			components = append(components, component)
		}
	}

	return components
}

// ============================================================================
// EXAMPLES AND DEMONSTRATION
// ============================================================================

// CreateExampleGraph creates a sample graph for demonstration
func CreateExampleGraph() *Graph {
	graph := NewGraph()

	// Add edges to create this graph:
	//     1 -- 2 -- 3
	//     |    |    |
	//     4 -- 5 -- 6
	//     |    |    |
	//     7 -- 8 -- 9

	edges := [][]int{
		{1, 2}, {1, 4}, {2, 3}, {2, 5}, {3, 6},
		{4, 5}, {4, 7}, {5, 6}, {5, 8}, {6, 9},
		{7, 8}, {8, 9},
	}

	for _, edge := range edges {
		graph.AddEdge(edge[0], edge[1])
	}

	return graph
}

// DemonstrateAlgorithms demonstrates different algorithms on the example graph
func DemonstrateAlgorithms() {
	fmt.Println("=== GRAPH TRAVERSAL ALGORITHMS DEMONSTRATION ===")

	// Create example graph
	graph := CreateExampleGraph()
	startNode := 1

	fmt.Println("Graph structure:")
	for node, neighbors := range graph.adjacencyList {
		fmt.Printf("  %d -> %v\n", node, neighbors)
	}
	fmt.Println()

	// DFS with stack
	dfsResult := graph.DFSIterativeStack(startNode)
	fmt.Printf("DFS (iterative with stack): %v\n", dfsResult)

	// DFS recursive
	dfsRecResult := graph.DFSRecursive(startNode)
	fmt.Printf("DFS (recursive): %v\n", dfsRecResult)

	// BFS with queue
	bfsResult := graph.BFSQueue(startNode)
	fmt.Printf("BFS (with queue): %v\n", bfsResult)

	fmt.Println("\n=== PATH FINDING ===")

	// Find path from 1 to 9
	dfsPath := graph.DFSFindPath(1, 9)
	bfsPath := graph.BFSShortestPath(1, 9)

	fmt.Printf("DFS path from 1 to 9: %v\n", dfsPath)
	fmt.Printf("BFS shortest path from 1 to 9: %v\n", bfsPath)

	fmt.Println("\n=== USE CASES ===")

	// Connected components
	components := graph.ConnectedComponentsDFS()
	fmt.Printf("Connected components: %v\n", components)

	// Cycle detection
	hasCycle := graph.DetectCycleDFS()
	fmt.Printf("Graph has cycle: %t\n", hasCycle)

	// Create a DAG for topological sort
	dag := NewGraph()
	dag.AddDirectedEdge(1, 2)
	dag.AddDirectedEdge(1, 3)
	dag.AddDirectedEdge(2, 4)
	dag.AddDirectedEdge(3, 4)
	dag.AddDirectedEdge(4, 5)

	topoOrder := dag.TopologicalSortDFS()
	fmt.Printf("Topological sort: %v\n", topoOrder)
}

func main() {
	DemonstrateAlgorithms()
}
