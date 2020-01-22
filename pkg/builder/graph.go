package builder

import "fmt"

// Node represents a node in the Graph
type Node struct {
	name string
	deps []string
}

// Graph holds all the nodes
type Graph struct {
	nodeNames map[string]*Node
}

// NewGraph returns a new initialize Graph structure
func NewGraph() Graph {
	return Graph{
		nodeNames: make(map[string]*Node),
	}
}

// AddNode add a node and its dependencies to the graph
func (g *Graph) AddNode(name string, deps ...string) {
	n := &Node{
		name: name,
		deps: deps,
	}
	if n.deps == nil {
		n.deps = []string{}
	}
	g.nodeNames[name] = n
}

// ResolveUpTo returns an ordered list of the dependencies required
// in other to fullfil finalNode
func (g *Graph) ResolveUpTo(finalNode string) ([]string, error) {
	if _, ok := g.nodeNames[finalNode]; !ok {
		return nil, fmt.Errorf("Final node doesn't exist in graph")
	}
	deps := reverseAndDeduplicate(g.dependenciesOf(finalNode))
	return append(deps, finalNode), nil
}

func (g *Graph) dependenciesOf(name string) []string {
	deps := []string{}
	for _, dep := range g.nodeNames[name].deps {
		deps = append(deps, g.nodeNames[dep].name)
		deps = append(deps, g.dependenciesOf(dep)...)
	}
	return deps
}

func reverseAndDeduplicate(g []string) []string {
	deps := []string{}
	exist := map[string]struct{}{}
	for i := 0; i < len(g); i++ {
		if _, ok := exist[g[len(g)-i-1]]; ok {
			continue
		}
		exist[g[len(g)-i-1]] = struct{}{}
		deps = append(deps, g[len(g)-i-1])
	}
	return deps
}
