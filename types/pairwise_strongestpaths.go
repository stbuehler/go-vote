package types

type StrongestPaths Pairwise

func (p PairwisePreferences) StrongestPaths() StrongestPaths {
	nChoices := len(p)
	paths := StrongestPaths(NewPairwise(nChoices))
	for i := 0; i < nChoices; i++ {
		for j := 0; j < nChoices; j++ {
			if i != j && p[i][j] > p[j][i] {
				paths[i][j] = p[i][j]
			}
		}
	}
	for i := 0; i < nChoices; i++ {
		for j := 0; j < nChoices; j++ {
			if i != j {
				for k := 0; k < nChoices; k++ {
					if i != k && j != k {
						// try path j -> i -> k instead of j -> k
						// new strength is minimum of j -> i and i -> k
						newStrength := paths[j][i]
						if newStrength > paths[i][k] {
							newStrength = paths[i][k]
						}
						// if new strength is better, take it
						if newStrength > paths[j][k] {
							paths[j][k] = newStrength
						}
					}
				}
			}
		}
	}
	return paths
}

func (p StrongestPaths) AsciiTable(candidates []string) string {
	return Pairwise(p).AsciiTable("p", candidates)
}

func (p StrongestPaths) Winner() int {
	return Pairwise(p).Winner()
}
