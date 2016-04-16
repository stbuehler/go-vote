package types

import (
	"errors"
)

var ErrRankOutOfRange = errors.New("Ranking contained negative rank")
var ErrMissingRank = errors.New("Ranking was not compact")
var ErrUnexpectedNumberOfCandidates = errors.New("RangGroups contained an unexpected number of candidates")
var ErrDuplicateCandidate = errors.New("RangGroups contained a candidate twice")
var ErrCandidateOutOfRange = errors.New("RangGroups contained a negative candidate")
var ErrCandidateMissing = errors.New("RangGroups misses a candidate")
var ErrEmptyRankGroup = errors.New("RangGroups contained an empty group")

/* for each candidate specify a "rank". each rank must be >= 0,
 * and all ranks between 0 and the highest rank must be used (i.e.
 * the ranking must be compact).
 *
 * smaller rank value means higher preference
 */

type Ranking []int

func (r Ranking) Check() error {
	if 0 == len(r) {
		return nil
	}
	highRank := 0
	for _, rank := range r {
		if rank < 0 {
			return ErrRankOutOfRange
		} else if highRank < rank {
			highRank = rank
		}
	}
	if highRank > len(r) {
		return ErrMissingRank
	}
	numUsedRanks := 0
	usedRank := make([]bool, highRank+1)
	for _, rank := range r {
		if !usedRank[rank] {
			usedRank[rank] = true
			numUsedRanks++
		}
	}
	if numUsedRanks != len(usedRank) {
		return ErrMissingRank
	}
	return nil
}

func (r Ranking) RankGroups() (RankGroups, error) {
	highRank := 0
	for _, rank := range r {
		if rank < 0 {
			return nil, ErrRankOutOfRange
		}
		if highRank < rank {
			highRank = rank
		}
	}
	if highRank > len(r) {
		return nil, ErrMissingRank
	}
	rankCountsOrOffset := make([]int, highRank+1)
	for _, rank := range r {
		rankCountsOrOffset[rank]++
	}
	groupsMemory := make([]int, len(r))
	rg := make(RankGroups, highRank+1)
	{
		offset := 0
		for rank, count := range rankCountsOrOffset {
			if 0 == count {
				return nil, ErrMissingRank
			}
			rg[rank] = groupsMemory[offset : offset+count]
			rankCountsOrOffset[rank] = offset // convert from count to offset
			offset += count
		}
	}
	for candidate, rank := range r {
		groupsMemory[rankCountsOrOffset[rank]] = candidate
		rankCountsOrOffset[rank]++
	}
	return rg, nil
}

/* for each rank in a `Ranking` this contains the list of all candidates
 * of that rank. similar to `Ranking` this must be compact, i.e. all
 * inner lists must be non-empty. Also all candidates [0..numCandidates[
 * must be present exactly once.
 */
type RankGroups [][]int

func (rg RankGroups) Sanitize(numCandidates int) RankGroups {
	mem := make([]int, 0, numCandidates)
	have := make([]bool, numCandidates)
	var newRankedGroups RankGroups
	for _, g := range rg {
		for _, candidate := range g {
			if 0 <= candidate && candidate < numCandidates && !have[candidate] {
				have[candidate] = true
				mem = append(mem, candidate)
			}
		}
		if 0 != len(mem) {
			newRankedGroups = append(newRankedGroups, mem)
			mem = mem[len(mem):]
		}
	}
	{
		for candidate := 0; candidate < numCandidates; candidate++ {
			if !have[candidate] {
				mem = append(mem, candidate)
			}
		}
		if 0 != len(mem) {
			newRankedGroups = append(newRankedGroups, mem)
		}
	}
	return newRankedGroups
}

func (rg RankGroups) Check(expectedNumCandidates int) error {
	_, err := rg.check(&expectedNumCandidates)
	return err
}

func (rg RankGroups) check(expectedNumCandidates *int) (numCandidates int, err error) {
	numCandidates = 0
	for _, g := range rg {
		if 0 == len(g) {
			return 0, ErrEmptyRankGroup
		}
		numCandidates += len(g)
	}
	if nil != expectedNumCandidates && *expectedNumCandidates != numCandidates {
		return 0, ErrUnexpectedNumberOfCandidates
	}
	have := make([]bool, numCandidates)
	for _, g := range rg {
		for _, candidate := range g {
			if candidate < 0 {
				return 0, ErrCandidateOutOfRange
			} else if candidate >= numCandidates {
				// index larger than should be according to number of
				// candidates: this means some candidate is not present
				// in the list
				return 0, ErrCandidateMissing
			} else if have[candidate] {
				return 0, ErrDuplicateCandidate
			}
			have[candidate] = true
		}
	}
	return numCandidates, nil
}

func (rg RankGroups) Ranking() (Ranking, error) {
	numCandidates := 0
	for _, g := range rg {
		if 0 == len(g) {
			return nil, ErrEmptyRankGroup
		}
		numCandidates += len(g)
	}
	r := make(Ranking, numCandidates)
	have := make([]bool, numCandidates)
	for rank, g := range rg {
		for _, candidate := range g {
			if candidate < 0 {
				return nil, ErrCandidateOutOfRange
			} else if candidate >= numCandidates {
				// index larger than should be according to number of
				// candidates: this means some candidate is not present
				// in the list
				return nil, ErrCandidateMissing
			} else if have[candidate] {
				return nil, ErrDuplicateCandidate
			}
			have[candidate] = true
			r[candidate] = rank
		}
	}
	return r, nil
}
