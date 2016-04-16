package backend

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stbuehler/go-vote/types"
	"log"
)

var ErrorUserNotFound = errors.New("User not found")
var ErrorInvalidUsername = errors.New("Invalid username")
var ErrorInvalidRanking = errors.New("Invalid ranking")
var ErrorElectionNotFound = errors.New("Election not found")
var ErrorElectionMembersOnly = errors.New("Only listed members can vote")
var ErrorElectionMembersOnlyEdit = errors.New("Only listed members can edit vote")
var ErrorElectionClosed = errors.New("Voting is closed")

type ElectionsTx struct {
	tx *sql.Tx
}

type User struct {
	Uid       int64
	Name      string
	Email     sql.NullString
	Token     sql.NullString
	SiteAdmin bool
}

type Election struct {
	Eid        int64
	Name       string // unique name identifier
	Title      string
	Candidates []string
	Closed     bool // whether election is closed
	Public     bool // whether unregistered/anonymous users can see election
	Open       bool // whether unregistered users can vote
	EditOpen   bool // whether votes from unregistered users can be edited
}

type Vote struct {
	Name    string
	Email   sql.NullString
	Ranking types.Ranking
}

func scanUser(row *sql.Row) (*User, error) {
	var user User
	if err := row.Scan(&user.Uid, &user.Name, &user.Email, &user.Token, &user.SiteAdmin); nil != err {
		return nil, err
	} else {
		if !user.Email.Valid || !user.Token.Valid {
			user.SiteAdmin = false
			user.Email = sql.NullString{}
			user.Token = sql.NullString{}
		}
		return &user, nil
	}
}

func (etx *ElectionsTx) Rollback() error {
	if nil == etx.tx {
		return nil
	} else {
		tx := etx.tx
		etx.tx = nil
		return tx.Rollback()
	}
}

func (etx *ElectionsTx) Commit() error {
	if nil == etx.tx {
		return nil
	} else {
		tx := etx.tx
		etx.tx = nil
		return tx.Commit()
	}
}

func (etx *ElectionsTx) FindUserByToken(token string) (*User, error) {
	row := etx.tx.QueryRow("SELECT uid, name, email, token, siteadmin FROM user WHERE token = ?", token)
	if user, err := scanUser(row); sql.ErrNoRows == err {
		return nil, ErrorUserNotFound
	} else if nil != err {
		log.Fatalf("FindUserByToken failed: %v", err)
		return nil, ErrorUserNotFound
	} else {
		return user, nil
	}
}

func (etx *ElectionsTx) FindOrCreateUnregisteredUser(name string) (*User, error) {
	if 0 == len(name) {
		return nil, ErrorInvalidUsername
	}
	row := etx.tx.QueryRow("SELECT uid, name, email, token, siteadmin FROM user WHERE name = ? AND email IS NULL", name)
	if user, err := scanUser(row); sql.ErrNoRows == err {
		etx.tx.Exec("INSERT INTO user (name) VALUES (?)", name)
		row := etx.tx.QueryRow("SELECT uid, name, email, token, siteadmin FROM user WHERE name = ? AND email IS NULL", name)
		if user, err := scanUser(row); nil != err {
			log.Fatalf("Couldn't add unregistered user %+q: %v", name, err)
			return nil, ErrorInvalidUsername
		} else {
			return user, nil
		}
	} else if nil != err {
		log.Fatalf("FindOrCreateUnregisteredUser failed: %v", err)
		return nil, ErrorInvalidUsername
	} else {
		return user, nil
	}
}

func scanElection(row *sql.Row) (*Election, error) {
	var e Election
	var candidatesJson string
	if err := row.Scan(&e.Eid, &e.Name, &e.Title, &candidatesJson, &e.Closed, &e.Public, &e.Open, &e.EditOpen); nil != err {
		return nil, err
	} else if err := json.Unmarshal([]byte(candidatesJson), &e.Candidates); nil != err {
		return nil, err
	} else {
		return &e, nil
	}
}

func (etx *ElectionsTx) CanSeeElection(user *User, e *Election) bool {
	if e.Public {
		return true
	}
	if nil == user {
		return false
	}
	if user.SiteAdmin {
		return true
	}
	var uid int64
	err := etx.tx.QueryRow("SELECT vote.uid FROM vote WHERE eid = ? AND uid = ?", e.Eid, user.Uid).Scan(&uid)
	return nil == err
}

func (etx *ElectionsTx) FindElectionByName(name string, user *User) *Election {
	row := etx.tx.QueryRow("SELECT eid, name, title, candidates, closed, public, open, editopen FROM election WHERE name = ?", name)
	if e, err := scanElection(row); sql.ErrNoRows == err {
		return nil
	} else if nil != err {
		log.Fatalf("FindElectionByName failed: %v", err)
		return nil
	} else if !etx.CanSeeElection(user, e) {
		return nil
	} else {
		return e
	}
}

func (etx *ElectionsTx) ElectionVotes(eid int64, offset, limit int) (int, []Vote, error) {
	var count int
	if err := etx.tx.QueryRow("SELECT COUNT(*) FROM vote WHERE eid = ?", eid).Scan(&count); nil != err {
		return 0, nil, fmt.Errorf("ElectionVotes count failed: %v", err)
	} else if offset >= count {
		return 0, nil, nil
	} else if limit > count-offset {
		limit = count - offset
	}
	if rows, err := etx.tx.Query("SELECT user.name, user.email, vote.ranking FROM vote LEFT JOIN user ON vote.uid = user.uid WHERE vote.eid = ? ORDER BY vote.uid LIMIT ? OFFSET ?", eid, limit, offset); nil != err {
		return 0, nil, fmt.Errorf("ElectionVotes failed: %v", err)
	} else {
		defer rows.Close()
		votes := make([]Vote, 0, limit)
		for rows.Next() {
			var v Vote
			var rankingJson sql.NullString
			if err := rows.Scan(&v.Name, &v.Email, &rankingJson); nil != err {
				return 0, nil, fmt.Errorf("ElectionVotes scan failed: %v", err)
			} else if rankingJson.Valid {
				if err := json.Unmarshal([]byte(rankingJson.String), &v.Ranking); nil != err {
					return 0, nil, fmt.Errorf("ElectionVotes parse ranking (%+q) failed: %v", rankingJson, err)
				}
			}
			votes = append(votes, v)
		}
		if err := rows.Err(); nil != err {
			return 0, nil, fmt.Errorf("ElectionVotes cursor failed: %v", err)
		}
		return count, votes, nil
	}
}

func (etx *ElectionsTx) ElectionPairwisePreferences(e *Election) (types.PairwisePreferences, error) {
	if rows, err := etx.tx.Query("SELECT ranking FROM vote WHERE vote.eid = ?", e.Eid); nil != err {
		return nil, fmt.Errorf("ElectionPairwisePreferences failed: %v", err)
	} else {
		defer rows.Close()
		numCandidates := len(e.Candidates)
		table := types.PairwisePreferences(types.NewPairwise(numCandidates))
		for rows.Next() {
			var rankingJson sql.NullString
			var ranking []int
			if err := rows.Scan(&rankingJson); nil != err {
				return nil, fmt.Errorf("ElectionPairwisePreferences scan failed: %v", err)
			} else if !rankingJson.Valid {
				continue
			} else if err := json.Unmarshal([]byte(rankingJson.String), &ranking); nil != err {
				return nil, fmt.Errorf("ElectionPairwisePreferences parse ranking (%+q) failed: %v", rankingJson.String, err)
			}
			if len(table) != numCandidates {
				return nil, fmt.Errorf("ElectionPairwisePreferences: inconsistent ranking lengths: %d != %d", numCandidates, len(ranking))
			}
			// count vote
			for runner := 0; runner < numCandidates; runner++ {
				for opponent := 0; opponent < numCandidates; opponent++ {
					if ranking[runner] < ranking[opponent] {
						table[runner][opponent]++
					}
				}
			}
		}
		if err := rows.Err(); nil != err {
			return nil, fmt.Errorf("ElectionPairwisePreferences cursor failed: %v", err)
		}
		return table, nil
	}
}

func (etx *ElectionsTx) CanVote(user *User, e *Election) error {
	var uid int64
	closedErr := error(nil)
	if e.Closed {
		// only leak "closed" information if all other checks were successful
		closedErr = ErrorElectionClosed
	}
	if nil == user {
		return ErrorElectionNotFound
	}
	if user.SiteAdmin {
		return closedErr
	}
	if e.Public && e.Open {
		if e.EditOpen || user.Email.Valid {
			return closedErr
		}
		// unregistered users can only vote if they didn't vote yet
		err := etx.tx.QueryRow("SELECT vote.uid FROM vote WHERE eid = ? AND uid = ?", e.Eid, user.Uid).Scan(&uid)
		if sql.ErrNoRows == err {
			return closedErr
		} else {
			return ErrorElectionMembersOnlyEdit
		}
	}
	// only listed and registered members can vote
	if !user.Email.Valid {
		return ErrorElectionMembersOnly
	}
	err := etx.tx.QueryRow("SELECT vote.uid FROM vote WHERE eid = ? AND uid = ?", e.Eid, user.Uid).Scan(&uid)
	if nil == err {
		return closedErr
	} else {
		return ErrorElectionMembersOnly
	}
}

func (etx *ElectionsTx) ElectionVote(e *Election, user *User, ranking types.Ranking) error {
	if err := etx.CanVote(user, e); nil != err {
		return err
	}
	if len(e.Candidates) != len(ranking) || nil != ranking.Check() {
		return ErrorInvalidRanking
	}
	rankingJson := types.JsonMustEncodeString(ranking)
	if !e.EditOpen && !user.Email.Valid {
		if _, err := etx.tx.Exec("INSERT INTO vote (eid, uid, ranking) VALUES (?, ?, ?)", e.Eid, user.Uid, rankingJson); nil != err {
			return ErrorElectionMembersOnlyEdit
		}
	} else {
		if _, err := etx.tx.Exec("INSERT OR REPLACE INTO vote (eid, uid, ranking) VALUES (?, ?, ?)", e.Eid, user.Uid, rankingJson); nil != err {
			log.Printf("Internal error when trying to insert vote: %v", err)
			return ErrorElectionNotFound
		}
	}
	log.Printf("Cast vote in election %d: user %d: %s", e.Eid, user.Uid, rankingJson)
	return nil
}
