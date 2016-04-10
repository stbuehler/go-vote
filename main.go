package main

import (
	"fmt"
	"github.com/stbuehler/go-vote/static"
	"net/http"
)

type testCase1 struct {
	NumVoters int
	Vote      Vote
}

var testCase1Data = []testCase1{
	testCase1{5, [][]int{{0}, {2}, {1}, {4}, {3}}},
	testCase1{5, [][]int{{0}, {3}, {4}, {2}, {1}}},
	testCase1{8, [][]int{{1}, {4}, {3}, {0}, {2}}},
	testCase1{3, [][]int{{2}, {0}, {1}, {4}, {3}}},
	testCase1{7, [][]int{{2}, {0}, {4}, {1}, {3}}},
	testCase1{2, [][]int{{2}, {1}, {0}, {3}, {4}}},
	testCase1{7, [][]int{{3}, {2}, {4}, {1}, {0}}},
	testCase1{8, [][]int{{4}, {1}, {0}, {3}, {2}}},
}

var testCase1Data2 = []testCase1{
	testCase1{1, [][]int{{0, 3}, {1, 2, 4}}},
	testCase1{1, [][]int{{1}, {0, 3, 2, 4}}},
	testCase1{1, [][]int{{2, 4}, {1, 0, 3}}},
}

func main() {
	var e Election
	e.Choices = []string{"A", "B", "C", "D", "E"}

	members := 0
	for _, row := range testCase1Data {
		for i := 0; i < row.NumVoters; i++ {
			members++
			memberName := fmt.Sprintf("member %d", members)
			e.Members = append(e.Members, memberName)
			e.Vote(memberName, row.Vote)
			// print(memberName + " voted:\n")
			// print(e.StringifyVote(row.Vote))
		}
	}

	/*
		prefs := e.Preferences()
		if winner := prefs.Winner(); -1 != winner {
			print("The winner is: " + e.Choices[winner] + "\n")
			print("\nThe pairwise preference are:\n")
			print(prefs.AsciiTable(&e))
		} else {
			print("There is no condorcet winner.\n")

			paths := prefs.StrongestPaths()

			if winner := paths.Winner(); -1 != winner {
				print("Schulze methode: the winner is: " + e.Choices[winner] + "\n")
			} else {
				print("Schulze methode: There is no winner.\n")
			}

			print("\nThe pairwise preference are:\n")
			print(prefs.AsciiTable(&e))

			print("\nThe strongest path strengths are:\n")
			print(paths.AsciiTable(&e))
		}
	*/

	api := &API{
		Elections: make(map[string]*Election),
	}
	api.Elections["2"] = &e
	{
		e := &Election{
			Choices: []string{"Red", "Green", "Blue", "Violet", "Yellow", "Black", "White"},
		}
		api.Elections["1"] = e
	}

	mux := http.NewServeMux()
	mux.Handle("/e/", api)
	mux.Handle("/", static.Handler)
	http.ListenAndServe(":8080", mux)
}
