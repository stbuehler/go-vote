package types

func TarjanSCC(edges [][]int) (mapping []int, components [][]int) {
	nodeCount := len(edges)

	var stack []int
	nodeOnStack := make([]bool, nodeCount)
	push := func(node int) {
		stack = append(stack, node)
		nodeOnStack[node] = true
	}
	pop := func() int {
		ndx := len(stack) - 1
		res := stack[ndx]
		stack = stack[:ndx]
		nodeOnStack[res] = false
		return res
	}
	min := func(a, b int) int {
		if a < b {
			return a
		} else {
			return b
		}
	}

	mapping = make([]int, nodeCount)

	nodeIndexes := make([]int, nodeCount)
	nodeLowLink := make([]int, nodeCount)
	index := 1
	var strongConnect func(node int)
	strongConnect = func(node int) {
		nodeIndexes[node] = index
		nodeLowLink[node] = index
		index++
		push(node)

		for _, link := range edges[node] {
			if 0 == nodeIndexes[link] {
				strongConnect(link)
				nodeLowLink[node] = min(nodeLowLink[node], nodeLowLink[link])
			} else if nodeOnStack[link] {
				nodeLowLink[node] = min(nodeLowLink[node], nodeIndexes[link])
			}
		}

		if nodeIndexes[node] == nodeLowLink[node] {
			var scc []int
			sccIndex := len(components)
			for {
				link := pop()
				scc = append(scc, link)
				mapping[link] = sccIndex
				if link == node {
					break
				}
			}
			components = append(components, scc)
		}
	}

	for node := 0; node < nodeCount; node++ {
		if 0 == nodeIndexes[node] {
			strongConnect(node)
		}
	}

	return
}
