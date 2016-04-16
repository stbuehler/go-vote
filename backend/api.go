package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stbuehler/go-vote/types"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var ApiInternalError = errors.New("Internal Server Error")

func apiInternalError() (int, interface{}, error) {
	return 500, nil, ApiInternalError
}

func apiInvalidRequest(err error) (int, interface{}, error) {
	return 400, nil, fmt.Errorf("Invalid request: %v", err)
}

func apiUnauthorizedRequest(err error) (int, interface{}, error) {
	return 401, nil, fmt.Errorf("Unauthorized request: %v", err)
}

func apiNotFound(err error) (int, interface{}, error) {
	return 404, nil, fmt.Errorf("Not found: %v", err)
}

type auth struct {
	Token string `json:",omitempty"`
	Name  string `json:",omitempty"`
}

func (etx ElectionsTx) findAuth(a auth) (*User, error) {
	if 0 != len(a.Token) {
		return etx.FindUserByToken(a.Token)
	} else {
		return nil, nil
	}
}

func (etx ElectionsTx) findOrCreateAuth(a auth) (*User, error) {
	if 0 != len(a.Token) {
		return etx.FindUserByToken(a.Token)
	} else {
		return etx.FindOrCreateUnregisteredUser(a.Name)
	}
}

type voteReq struct {
	Auth       auth
	RankGroups types.RankGroups
}

func (edb ElectionsDb) apiHandleVote(query url.Values, jsonBody []byte) (int, interface{}, error) {
	var req voteReq
	if err := json.Unmarshal(jsonBody, &req); nil != err {
		return apiInvalidRequest(err)
	} else if etx, err := edb.StartTransaction(); nil != err {
		return apiInternalError()
	} else {
		defer etx.Rollback()

		if user, err := etx.findOrCreateAuth(req.Auth); nil != err {
			return apiUnauthorizedRequest(err)
		} else if e := etx.FindElectionByName(query.Get("election"), user); nil == e {
			return apiNotFound(fmt.Errorf("Election not found"))
		} else if err := req.RankGroups.Check(len(e.Candidates)); nil != err {
			return apiInvalidRequest(err)
		} else if ranking, err := req.RankGroups.Ranking(); nil != err {
			return apiInvalidRequest(err)
		} else if err := etx.ElectionVote(e, user, ranking); nil != err {
			return apiInvalidRequest(err)
		} else if err := etx.Commit(); nil != err {
			log.Printf("Vote commit failed: %v", err)
			return apiInternalError()
		} else {
			rankingJson := types.JsonMustEncodeString(ranking)
			log.Printf("Committed vote: eid=%d uid=%d ranking=%s", e.Eid, user.Uid, rankingJson)
			return 200, nil, nil
		}
	}
}

func (edb ElectionsDb) ApiVoteHandler() http.HandlerFunc {
	return makeApiHandler(edb.apiHandleVote)
}

type resultsReq struct {
	Auth auth
}

func (edb ElectionsDb) apiHandleResults(query url.Values, jsonBody []byte) (int, interface{}, error) {
	var req resultsReq
	if err := json.Unmarshal(jsonBody, &req); nil != err {
		return apiInvalidRequest(err)
	} else if etx, err := edb.StartTransaction(); nil != err {
		return apiInternalError()
	} else {
		defer etx.Rollback()

		if user, err := etx.findAuth(req.Auth); nil != err {
			return apiUnauthorizedRequest(err)
		} else if e := etx.FindElectionByName(query.Get("election"), user); nil == e {
			return apiNotFound(fmt.Errorf("Election not found"))
		} else if pairPrefs, err := etx.ElectionPairwisePreferences(e); nil != err {
			log.Printf("ApiResults failure: %v", err)
			return apiInternalError()
		} else if err := etx.Commit(); nil != err {
			log.Printf("ApiResults failure: %v", err)
			return apiInternalError()
		} else {
			result := make(map[string]interface{})
			result["preferences"] = pairPrefs
			if winner := pairPrefs.Winner(); -1 != winner {
				result["winner"] = winner
			} else {
				paths := pairPrefs.StrongestPaths()
				result["paths"] = paths
				if winner := paths.Winner(); -1 != winner {
					result["winner"] = winner
				}
			}

			return 200, result, nil
		}
	}
}

func (edb ElectionsDb) ApiResultsHandler() http.HandlerFunc {
	return makeApiHandler(edb.apiHandleResults)
}

func makeApiHandler(api func(query url.Values, jsonBody []byte) (int, interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		jsonBody, err := ioutil.ReadAll(req.Body)
		if nil != err {
			log.Printf("Couldn't read json request body: %v", err)
			http.Error(w, "400 Bad Request", 400)
			return
		} else if code, result, err := api(req.URL.Query(), jsonBody); nil != err {
			log.Printf("Request[%+q] failed: %d %+q", req.URL.EscapedPath(), code, err)
			http.Error(w, err.Error(), code)
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(code)
			w.Write(types.JsonMustEncode(result))
		}
	}
}

func (edb ElectionsDb) BindServeMux(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/vote", edb.ApiVoteHandler())
	mux.HandleFunc(prefix+"/result", edb.ApiResultsHandler())
}
