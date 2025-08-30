from collections import deque, defaultdict
from typing import List, Set, Dict, Optional

class Graph:
    def __init__(self):
        self.adjacency_list = defaultdict(list)
    
    def add_edge(self, u: int, v: int):
        """Add an undirected edge between nodes u and v"""
        self.adjacency_list[u].append(v)
        self.adjacency_list[v].append(u)
    
    def add_directed_edge(self, u: int, v: int):
        """Add a directed edge from u to v"""
        self.adjacency_list[u].append(v)


# ============================================================================
# DEPTH-FIRST SEARCH (DFS) WITH STACK
# ============================================================================

def dfs_iterative_stack(graph: Graph, start: int) -> List[int]:
    """
    DFS using explicit stack (iterative approach)
    Time Complexity: O(V + E)
    Space Complexity: O(V)
    """
    if not graph.adjacency_list:
        return []
    
    visited = set()
    stack = [start]
    result = []
    
    while stack:
        # Pop from top of stack (LIFO)
        current = stack.pop()
        
        if current not in visited:
            visited.add(current)
            result.append(current)
            
            # Push all unvisited neighbors onto stack
            # Note: We push in reverse order to maintain natural traversal order
            for neighbor in reversed(graph.adjacency_list[current]):
                if neighbor not in visited:
                    stack.append(neighbor)
    
    return result

def dfs_recursive(graph: Graph, start: int) -> List[int]:
    """
    DFS using recursion (implicit call stack)
    This is equivalent to using a stack but uses the system call stack
    """
    visited = set()
    result = []
    
    def dfs_helper(node: int):
        if node in visited:
            return
        
        visited.add(node)
        result.append(node)
        
        for neighbor in graph.adjacency_list[node]:
            if neighbor not in visited:
                dfs_helper(neighbor)
    
    dfs_helper(start)
    return result

# ============================================================================
# BREADTH-FIRST SEARCH (BFS) WITH QUEUE
# ============================================================================

def bfs_queue(graph: Graph, start: int) -> List[int]:
    """
    BFS using explicit queue
    Time Complexity: O(V + E)
    Space Complexity: O(V)
    """
    if not graph.adjacency_list:
        return []
    
    visited = set()
    queue = deque([start])
    result = []
    
    while queue:
        # Dequeue from front (FIFO)
        current = queue.popleft()
        
        if current not in visited:
            visited.add(current)
            result.append(current)
            
            # Enqueue all unvisited neighbors
            for neighbor in graph.adjacency_list[current]:
                if neighbor not in visited:
                    queue.append(neighbor)
    
    return result

# ============================================================================
# ADVANCED EXAMPLES WITH PATH FINDING
# ============================================================================

def dfs_find_path(graph: Graph, start: int, end: int) -> Optional[List[int]]:
    """
    DFS to find a path from start to end node
    Returns the path if found, None otherwise
    """
    if start == end:
        return [start]
    
    visited = set()
    stack = [(start, [start])]  # (node, path_to_node)
    
    while stack:
        current, path = stack.pop()
        
        if current == end:
            return path
        
        if current not in visited:
            visited.add(current)
            
            for neighbor in graph.adjacency_list[current]:
                if neighbor not in visited:
                    new_path = path + [neighbor]
                    stack.append((neighbor, new_path))
    
    return None

def bfs_shortest_path(graph: Graph, start: int, end: int) -> Optional[List[int]]:
    """
    BFS to find shortest path from start to end node
    Returns the shortest path if found, None otherwise
    """
    if start == end:
        return [start]
    
    visited = set()
    queue = deque([(start, [start])])  # (node, path_to_node)
    
    while queue:
        current, path = queue.popleft()
        
        if current == end:
            return path
        
        if current not in visited:
            visited.add(current)
            
            for neighbor in graph.adjacency_list[current]:
                if neighbor not in visited:
                    new_path = path + [neighbor]
                    queue.append((neighbor, new_path))
    
    return None

# ============================================================================
# USE CASES AND APPLICATIONS
# ============================================================================

def detect_cycle_dfs(graph: Graph) -> bool:
    """
    Detect cycle in undirected graph using DFS
    """
    visited = set()
    
    def has_cycle_dfs(node: int, parent: int) -> bool:
        visited.add(node)
        
        for neighbor in graph.adjacency_list[node]:
            if neighbor not in visited:
                if has_cycle_dfs(neighbor, node):
                    return True
            elif neighbor != parent:
                return True  # Back edge found
        return False
    
    for node in graph.adjacency_list:
        if node not in visited:
            if has_cycle_dfs(node, -1):
                return True
    return False

def topological_sort_dfs(graph: Graph) -> List[int]:
    """
    Topological sort using DFS (for directed acyclic graphs)
    """
    visited = set()
    temp_visited = set()  # For cycle detection
    result = []
    
    def dfs_topo(node: int) -> bool:
        if node in temp_visited:
            return False  # Cycle detected
        if node in visited:
            return True
        
        temp_visited.add(node)
        
        for neighbor in graph.adjacency_list[node]:
            if not dfs_topo(neighbor):
                return False
        
        temp_visited.remove(node)
        visited.add(node)
        result.append(node)
        return True
    
    for node in graph.adjacency_list:
        if node not in visited:
            if not dfs_topo(node):
                return []  # Cycle detected
    
    return result[::-1]  # Reverse to get topological order

def connected_components_dfs(graph: Graph) -> List[List[int]]:
    """
    Find all connected components using DFS
    """
    visited = set()
    components = []
    
    def dfs_component(node: int, component: List[int]):
        visited.add(node)
        component.append(node)
        
        for neighbor in graph.adjacency_list[node]:
            if neighbor not in visited:
                dfs_component(neighbor, component)
    
    for node in graph.adjacency_list:
        if node not in visited:
            component = []
            dfs_component(node, component)
            components.append(component)
    
    return components

# ============================================================================
# EXAMPLES AND DEMONSTRATION
# ============================================================================

def create_example_graph() -> Graph:
    """Create a sample graph for demonstration"""
    graph = Graph()
    
    # Add edges to create this graph:
    #     1 -- 2 -- 3
    #     |    |    |
    #     4 -- 5 -- 6
    #     |    |    |
    #     7 -- 8 -- 9
    
    edges = [
        (1, 2), (1, 4), (2, 3), (2, 5), (3, 6),
        (4, 5), (4, 7), (5, 6), (5, 8), (6, 9),
        (7, 8), (8, 9)
    ]
    
    for u, v in edges:
        graph.add_edge(u, v)
    
    return graph

def demonstrate_algorithms():
    """Demonstrate different algorithms on the example graph"""
    print("=== GRAPH TRAVERSAL ALGORITHMS DEMONSTRATION ===\n")
    
    # Create example graph
    graph = create_example_graph()
    start_node = 1
    
    print(f"Graph structure:")
    for node, neighbors in graph.adjacency_list.items():
        print(f"  {node} -> {neighbors}")
    print()
    
    # DFS with stack
    dfs_result = dfs_iterative_stack(graph, start_node)
    print(f"DFS (iterative with stack): {dfs_result}")
    
    # DFS recursive
    dfs_rec_result = dfs_recursive(graph, start_node)
    print(f"DFS (recursive): {dfs_rec_result}")
    
    # BFS with queue
    bfs_result = bfs_queue(graph, start_node)
    print(f"BFS (with queue): {bfs_result}")
    
    print("\n=== PATH FINDING ===")
    
    # Find path from 1 to 9
    dfs_path = dfs_find_path(graph, 1, 9)
    bfs_path = bfs_shortest_path(graph, 1, 9)
    
    print(f"DFS path from 1 to 9: {dfs_path}")
    print(f"BFS shortest path from 1 to 9: {bfs_path}")
    
    print("\n=== USE CASES ===")
    
    # Connected components
    components = connected_components_dfs(graph)
    print(f"Connected components: {components}")
    
    # Cycle detection
    has_cycle = detect_cycle_dfs(graph)
    print(f"Graph has cycle: {has_cycle}")
    
    # Create a DAG for topological sort
    dag = Graph()
    dag.add_directed_edge(1, 2)
    dag.add_directed_edge(1, 3)
    dag.add_directed_edge(2, 4)
    dag.add_directed_edge(3, 4)
    dag.add_directed_edge(4, 5)
    
    topo_order = topological_sort_dfs(dag)
    print(f"Topological sort: {topo_order}")

if __name__ == "__main__":
    demonstrate_algorithms()
