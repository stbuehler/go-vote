package types

type PairwisePreferences Pairwise

func (p PairwisePreferences) AsciiTable(candidates []string) string {
	return Pairwise(p).AsciiTable("d", candidates)
}

func (p PairwisePreferences) Winner() int {
	return Pairwise(p).Winner()
}
