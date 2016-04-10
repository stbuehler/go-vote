package main

import (
	"bytes"
	"errors"
	"strings"
)

var AccessDenied = errors.New("Access denied")

type Election struct {
	Choices []string
	Votes   map[string]Vote
	Members []string
	Admins  []string
}

type Vote [][]int

// entry[x][y]: how often x was preferred over y
type Preferences [][]int
type StrongestPaths [][]int

func (e *Election) SanitizeVote(v Vote) Vote {
	nChoices := len(e.Choices)
	have := make([]bool, nChoices)
	var vote Vote
	for _, rank := range v {
		var sRank []int
		for _, choice := range rank {
			if 0 <= choice && choice < nChoices && !have[choice] {
				have[choice] = true
				sRank = append(sRank, choice)
			}
		}
		if 0 != len(sRank) {
			vote = append(vote, sRank)
		}
	}
	{
		var sRank []int
		for i := 0; i < nChoices; i++ {
			if !have[i] {
				sRank = append(sRank, i)
			}
		}
		if 0 != len(sRank) {
			vote = append(vote, sRank)
		}
	}
	return vote
}

func (e *Election) IsMember(member string) bool {
	for _, m := range e.Members {
		if m == member {
			return true
		}
	}
	return false
}

func (e *Election) IsAdmin(admin string) bool {
	for _, a := range e.Admins {
		if a == admin {
			return true
		}
	}
	return false
}

func (e *Election) Vote(member string, v Vote) error {
	if !e.IsMember(member) {
		return AccessDenied
	}
	if nil == e.Votes {
		e.Votes = make(map[string]Vote)
	}
	e.Votes[member] = e.SanitizeVote(v)
	return nil
}

func alloc2DimInt(cols, rows int) [][]int {
	cells := make([]int, cols*rows)
	table := make([][]int, rows)
	for r := 0; r < rows; r++ {
		table[r], cells = cells[:cols], cells[cols:]
	}
	return table
}

func (e *Election) Preferences() Preferences {
	nChoices := len(e.Choices)
	prefs := Preferences(alloc2DimInt(nChoices, nChoices))
	for _, vote := range e.Votes {
		vote.AddToPreferences(prefs)
	}
	return prefs
}

func (e *Election) ParseVote(text string) Vote {
	var vote Vote
	for _, line := range strings.Split(text, "\n") {
		var rank []int
		for _, choice := range strings.Split(line, ",") {
			choice = strings.TrimSpace(choice)
			if 0 == len(choice) {
				continue
			}
			for i, c := range e.Choices {
				if strings.EqualFold(c, choice) {
					rank = append(rank, i)
					break
				}
			}
		}
		if 0 != len(rank) {
			vote = append(vote, rank)
		}
	}
	return vote
}

func (e *Election) StringifyVote(v Vote) string {
	var buf bytes.Buffer
	for _, rank := range v {
		for i, choice := range rank {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(e.Choices[choice])
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func (v Vote) AddToPreferences(prefs Preferences) {
	preferred := make([]int, 0, len(prefs))
	for _, rank := range v {
		for _, looser := range rank {
			for _, winner := range preferred {
				prefs[winner][looser]++
			}
		}
		preferred = append(preferred, rank...)
	}
}

func (p Preferences) AsciiTable(e *Election) string {
	return AsciiIntTable([][]int(p), "d", e.Choices)
}

func tableWinner(table [][]int) int {
	nChoices := len(table)
nextCandidate:
	for i := 0; i < nChoices; i++ {
		for j := 0; j < nChoices; j++ {
			// if i didn't win over j more often than it lost it doesn't win
			if i != j && table[i][j] <= table[j][i] {
				continue nextCandidate
			}
		}
		// won all pairwise matches
		return i
	}
	return -1
}

func (p Preferences) Winner() int {
	return tableWinner([][]int(p))
}

func (p Preferences) StrongestPaths() StrongestPaths {
	nChoices := len(p)
	paths := StrongestPaths(alloc2DimInt(nChoices, nChoices))
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

func (p StrongestPaths) AsciiTable(e *Election) string {
	return AsciiIntTable([][]int(p), "p", e.Choices)
}

func (p StrongestPaths) Winner() int {
	return tableWinner([][]int(p))
}
