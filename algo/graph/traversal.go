package graph

import (
	"container/list"
	"fmt"
)

// AdjGraph represents a graph using an adjacency list.
// Set Directed to true to treat edges as one-way.
type AdjGraph struct {
	Adjacency map[int][]int
	Directed  bool
}

// NewAdjGraph creates a new adjacency-list graph.
func NewAdjGraph(directed bool) *AdjGraph {
	return &AdjGraph{Adjacency: make(map[int][]int), Directed: directed}
}

// AddEdge adds an edge between u and v. If the graph is undirected, it adds both directions.
func (g *AdjGraph) AddEdge(u, v int) {
	g.Adjacency[u] = append(g.Adjacency[u], v)
	if !g.Directed {
		g.Adjacency[v] = append(g.Adjacency[v], u)
	}
}

// BFS performs breadth-first traversal starting from start and returns the visit order.
func (g *AdjGraph) BFS(start int) []int {
	visited := make(map[int]bool)
	order := []int{}
	q := list.New()
	q.PushBack(start)

	for q.Len() > 0 {
		n := q.Remove(q.Front()).(int)
		if visited[n] {
			continue
		}
		visited[n] = true
		order = append(order, n)
		for _, nb := range g.Adjacency[n] {
			if !visited[nb] {
				q.PushBack(nb)
			}
		}
	}
	return order
}

// DFSIterative performs depth-first traversal using an explicit stack.
func (g *AdjGraph) DFSIterative(start int) []int {
	visited := make(map[int]bool)
	order := []int{}
	stack := list.New()
	stack.PushBack(start)

	for stack.Len() > 0 {
		n := stack.Remove(stack.Back()).(int)
		if visited[n] {
			continue
		}
		visited[n] = true
		order = append(order, n)
		// Push neighbors in reverse order to get a natural left-to-right feel
		nbs := g.Adjacency[n]
		for i := len(nbs) - 1; i >= 0; i-- {
			nb := nbs[i]
			if !visited[nb] {
				stack.PushBack(nb)
			}
		}
	}
	return order
}

// DFSRecursive performs depth-first traversal using recursion.
func (g *AdjGraph) DFSRecursive(start int) []int {
	visited := make(map[int]bool)
	order := []int{}

	var dfs func(int)
	dfs = func(n int) {
		if visited[n] {
			return
		}
		visited[n] = true
		order = append(order, n)
		for _, nb := range g.Adjacency[n] {
			if !visited[nb] {
				dfs(nb)
			}
		}
	}

	dfs(start)
	return order
}

// ShortestPath returns the shortest path in an unweighted graph using BFS (inclusive of start and end).
// Returns nil if no path exists.
func (g *AdjGraph) ShortestPath(start, end int) []int {
	if start == end {
		return []int{start}
	}
	visited := make(map[int]bool)
	prev := make(map[int]int)
	q := list.New()
	q.PushBack(start)

	for q.Len() > 0 {
		n := q.Remove(q.Front()).(int)
		if visited[n] {
			continue
		}
		visited[n] = true
		for _, nb := range g.Adjacency[n] {
			if !visited[nb] {
				prev[nb] = n
				if nb == end {
					// reconstruct
					path := []int{end}
					for cur := end; cur != start; {
						p, ok := prev[cur]
						if !ok {
							return nil
						}
						path = append(path, p)
						cur = p
					}
					// reverse
					for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
						path[i], path[j] = path[j], path[i]
					}
					return path
				}
				q.PushBack(nb)
			}
		}
	}
	return nil
}

// HasCycleUndirected detects a cycle in an undirected graph using DFS.
func (g *AdjGraph) HasCycleUndirected() bool {
	visited := make(map[int]bool)

	var dfs func(node, parent int) bool
	dfs = func(node, parent int) bool {
		visited[node] = true
		for _, nb := range g.Adjacency[node] {
			if !visited[nb] {
				if dfs(nb, node) {
					return true
				}
			} else if nb != parent {
				return true
			}
		}
		return false
	}

	for n := range g.Adjacency {
		if !visited[n] {
			if dfs(n, -1) {
				return true
			}
		}
	}
	return false
}

// TopoSortKahn computes a topological order using Kahn's algorithm.
// It returns the order and a boolean indicating whether the graph is a DAG.
func (g *AdjGraph) TopoSortKahn() ([]int, bool) {
	if !g.Directed {
		return nil, false
	}
	// compute in-degrees
	indeg := make(map[int]int)
	for u := range g.Adjacency {
		if _, ok := indeg[u]; !ok {
			indeg[u] = 0
		}
		for _, v := range g.Adjacency[u] {
			indeg[v]++
		}
	}
	q := list.New()
	for n, d := range indeg {
		if d == 0 {
			q.PushBack(n)
		}
	}
	order := []int{}
	for q.Len() > 0 {
		u := q.Remove(q.Front()).(int)
		order = append(order, u)
		for _, v := range g.Adjacency[u] {
			indeg[v]--
			if indeg[v] == 0 {
				q.PushBack(v)
			}
		}
	}
	if len(order) != len(indeg) {
		return order, false // cycle exists
	}
	return order, true
}

// ConnectedComponents returns all connected components (for undirected graphs) using DFS.
func (g *AdjGraph) ConnectedComponents() [][]int {
	visited := make(map[int]bool)
	res := [][]int{}

	var dfs func(int, *[]int)
	dfs = func(n int, comp *[]int) {
		visited[n] = true
		*comp = append(*comp, n)
		for _, nb := range g.Adjacency[n] {
			if !visited[nb] {
				dfs(nb, comp)
			}
		}
	}

	for n := range g.Adjacency {
		if !visited[n] {
			comp := []int{}
			dfs(n, &comp)
			res = append(res, comp)
		}
	}
	return res
}

// Demo prints traversal orders and common graph-ops using a small sample graph.
func Demo() {
	g := NewAdjGraph(false)
	for _, e := range [][2]int{{1, 2}, {1, 4}, {2, 3}, {2, 5}, {3, 6}, {4, 5}, {4, 7}, {5, 6}, {5, 8}, {6, 9}, {7, 8}, {8, 9}} {
		g.AddEdge(e[0], e[1])
	}

	fmt.Println("BFS from 1:", g.BFS(1))
	fmt.Println("DFS (iterative) from 1:", g.DFSIterative(1))
	fmt.Println("DFS (recursive) from 1:", g.DFSRecursive(1))
	fmt.Println("Shortest path 1->9:", g.ShortestPath(1, 9))
	fmt.Println("Has cycle (undirected):", g.HasCycleUndirected())
	fmt.Println("Connected components:", g.ConnectedComponents())

	dag := NewAdjGraph(true)
	for _, e := range [][2]int{{1, 2}, {1, 3}, {2, 4}, {3, 4}, {4, 5}} {
		dag.AddEdge(e[0], e[1])
	}
	order, ok := dag.TopoSortKahn()
	fmt.Println("Topo order:", order, "DAG:", ok)
}
