package frontend

import (
	"fmt"
	"github.com/stbuehler/go-vote/backend"
	"github.com/stbuehler/go-vote/static"
	"github.com/stbuehler/go-vote/types"
	"net/http"
)

type Frontend struct {
	Edb backend.ElectionsDb
}

func (f Frontend) BindServeMux(mux *http.ServeMux, prefix string) {
	path := prefix + "/e/"

	pathSortableJS := static.PathSortableJS(prefix)
	pathVoteJS := static.PathVoteJS(prefix)
	pathApiJS := static.PathApiJS(prefix)
	pathVoteCSS := static.PathVoteCSS(prefix)

	mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		electionName := req.URL.Path[len(path):]
		if etx, err := f.Edb.StartTransaction(); nil != err {
			http.Error(w, "Internal server error", 500)
		} else {
			defer etx.Rollback()

			if e := etx.FindElectionByName(electionName, nil); nil == e {
				http.Error(w, "Election not found", 404)
			} else {
				w.Header().Add("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Vote</title>

  <script src="%s"></script>
  <script src="%s"></script>
  <script src="%s"></script>
  <link rel="stylesheet" href="%s">
</head>
<body style="text-align: center;">
  <div style="display: inline-block; text-align: left;">
    <div class="block" id="vote-block" style="max-width: 400px; margin: 10px; display: inline-block;">
      <h2>Rank according to your preferences</h2>
      <p>Choices in the same block have equal preference. Choices in blocks at the top are preferred over choices in lower blocks.</p>
      <div id="vote"></div>
      <p><label>Name: <input id="voter" type="text" size="30"></input></label></p>
      <p><button id="submit-vote">Submit</button></p>
    </div>
    <div class="block" id="result-block" style="max-width: 400px; margin: 10px; display: inline-block;">
      <h2>Result</h2>
      <p><button id="submit-result">Reload</button></p>
      <div id="result"></div>
    </div>
  </div>
  <script>//<![CDATA[

(function() {
  var prefix = %s;
  var electionName = %s;
  var choices = %s;
  var rankGroups = %s;
  setup(prefix, electionName, choices, rankGroups);
})();

  //]]></script>
</body>
</html>`,
					pathSortableJS,
					pathVoteJS,
					pathApiJS,
					pathVoteCSS,
					types.JsonMustEncodeString(prefix),
					types.JsonMustEncodeString(electionName),
					types.JsonMustEncodeString(e.Candidates),
					types.JsonMustEncodeString(types.RankGroups{}.Sanitize(len(e.Candidates))),
				)
			}
		}
	})
}
