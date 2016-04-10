package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type API struct {
	Elections map[string]*Election
}

func (a *API) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if path[:3] != "/e/" || 3 == len(path) {
		http.NotFound(w, req)
		return
	}
	var baseURL string
	var action string
	if l := strings.SplitN(path[3:], "/", 3); len(l) > 2 {
		log.Printf("Invalid path: %+q", path)
		http.Error(w, "400 Bad Request", 400)
		return
	} else if 1 == len(l) {
		action, path = "view", path[3:]
		baseURL = "../"
	} else {
		action, path = l[0], l[1]
		baseURL = "../../"
	}

	e := a.Elections[path]
	if nil == e {
		log.Printf("Election not found: %+q", path)
		http.NotFound(w, req)
		return
	}

	switch action {
	case "view":
		e.ViewHTML(baseURL, path, w)
	case "vote":
		type VoteData struct {
			Voter string
			Vote  Vote
		}
		var data VoteData
		body, err := ioutil.ReadAll(req.Body)
		if nil != err {
			log.Printf("Couldn't read 'vote' request body: %v", err)
			http.Error(w, "400 Bad Request", 400)
			return
		}
		if err := json.Unmarshal(body, &data); nil != err {
			log.Printf("Couldn't parse 'vote' request body: %v (%+q), %v", err, string(body), req)
			http.Error(w, "400 Bad Request", 400)
			return
		}
		if e.IsMember(data.Voter) || 0 == len(data.Voter) {
			log.Printf("Already voted or invalid name: %+q", data.Voter)
			http.Error(w, "409 Conflict", 409)
			return
		}
		e.Members = append(e.Members, data.Voter)
		e.Vote(data.Voter, data.Vote)
	case "result":
		result := make(map[string]interface{})
		prefs := e.Preferences()
		result["preferences"] = prefs
		if winner := prefs.Winner(); -1 != winner {
			result["winner"] = winner
		} else {
			paths := prefs.StrongestPaths()
			result["paths"] = paths
			if winner := paths.Winner(); -1 != winner {
				result["winner"] = winner
			}
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(JsonMustEncode(result))
	default:
		http.NotFound(w, req)
	}
}
