package types

import (
	"bytes"
	"fmt"
	"sort"
)

// entries: [runner][oppenent]
// usually one runner is shown in one line: [y][x], [row][column]
type Pairwise [][]int

func NewPairwise(numCandidates int) Pairwise {
	return Pairwise(Alloc2DimInt(numCandidates, numCandidates))
}

func (table Pairwise) Winner() int {
	numCandidates := len(table)
nextCandidate:
	for runner := 0; runner < numCandidates; runner++ {
		for opponent := 0; opponent < numCandidates; opponent++ {
			// if runner didn't win over opponent more often than it lost it doesn't win
			if runner != opponent && table[runner][opponent] <= table[opponent][runner] {
				continue nextCandidate
			}
		}
		// won all pairwise matches
		return runner
	}
	return -1
}

type sortComponentsByMostEdges struct {
	components [][]int
	edgeCount  []int
}

func (s sortComponentsByMostEdges) Len() int {
	return len(s.edgeCount)
}
func (s sortComponentsByMostEdges) Less(i, j int) bool {
	return s.edgeCount[i] > s.edgeCount[j]
}
func (s sortComponentsByMostEdges) Swap(i, j int) {
	s.edgeCount[i], s.edgeCount[j] = s.edgeCount[j], s.edgeCount[i]
	s.components[i], s.components[j] = s.components[j], s.components[i]
}

func (table Pairwise) Ranking() Ranking {
	numCandidates := len(table)
	edgesMemory := make([]int, 0, numCandidates*numCandidates)
	edges := make([][]int, numCandidates)

	for runner := 0; runner < numCandidates; runner++ {
		for opponent := 0; opponent < numCandidates; opponent++ {
			if runner != opponent && table[runner][opponent] >= table[opponent][runner] {
				// ties and wins get an edge
				edgesMemory = append(edgesMemory, opponent)
			}
		}
		// slice edges
		numEdges := len(edgesMemory)
		edges[runner], edgesMemory = edgesMemory[:numEdges], edgesMemory[numEdges:]
	}

	mapping, components := TarjanSCC(edges)

	/* for i != j there is an edge "i -> j" or "j -> i" (or both)
	 * if the strongly connected components get merged, the set of edge
	 * is a strict total order, i.e. we can simply sort the SCCs by the
	 * number of outgoing edges
	 */

	numComps := len(components)
	// sccEdgeCount[i]: outgoing edges for "SCC i"
	sccEdgeCount := make([]int, numComps)
	{
		/* count merged edges for the merged components; make sure to
		 * not count edges twice
		 *
		 * sccHaveEdge[i][j]: whether there is an edge "SCC i -> SCC j"
		 */
		sccHaveEdge := make([][]bool, numComps)
		{
			mem := make([]bool, numComps*numComps)
			for ndx := 0; ndx < numComps; ndx++ {
				sccHaveEdge[ndx], mem = mem[:numComps], mem[numComps:]
			}
		}
		for from, links := range edges {
			mappedFrom := mapping[from]
			for _, to := range links {
				mappedTo := mapping[to]
				if !sccHaveEdge[mappedFrom][mappedTo] {
					sccHaveEdge[mappedFrom][mappedTo] = true
					sccEdgeCount[mappedFrom]++
				}
			}
		}
	}
	// now sort components by outgoing edges. the first component should
	// have the most outgoing edges.
	sort.Sort(sortComponentsByMostEdges{
		components: components,
		edgeCount:  sccEdgeCount,
	})

	ranking := make(Ranking, numCandidates)
	for rank, component := range components {
		for _, candidate := range component {
			ranking[candidate] = rank
		}
	}
	return ranking
}

func (table Pairwise) AsciiTable(tableName string, labels []string) string {
	var buf bytes.Buffer
	colSize := 0
	for i, _ := range table {
		if colSize < len(labels[i]) {
			colSize = len(labels[i])
		}
	}
	colSize = colSize + len(tableName) + 5
	colFormat := fmt.Sprintf("%%%ds |", colSize)
	entryFormat := fmt.Sprintf("%%%dd |", colSize)

	fmt.Fprint(&buf, "|")
	fmt.Fprintf(&buf, colFormat, tableName)
	for i, _ := range table {
		fmt.Fprintf(&buf, colFormat, fmt.Sprintf("%s[*,%s]", tableName, labels[i]))
	}
	fmt.Fprint(&buf, "\n")

	for i, line := range table {
		fmt.Fprint(&buf, "|")
		fmt.Fprintf(&buf, colFormat, fmt.Sprintf("%s[%s,*]", tableName, labels[i]))
		for _, wins := range line {
			fmt.Fprintf(&buf, entryFormat, wins)
		}
		fmt.Fprint(&buf, "\n")
	}
	return buf.String()
}
