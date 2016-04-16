package static

import (
	"fmt"
	"net/http"
)

func IndexForPrefix(prefix string) http.HandlerFunc {
	pathSortableJS := PathSortableJS(prefix)
	pathVoteJS := PathVoteJS(prefix)
	pathVoteCSS := PathVoteCSS(prefix)
	path := prefix + "/"

	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != path {
			http.NotFound(w, req)
		} else {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w,
				`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Vote</title>

  <script src="%s"></script>
  <script src="%s"></script>
  <link rel="stylesheet" href="%s">
</head>
<body>
  <div class="block" id="vote-block">
    <h2>Vote Example</h2>
    <div id="vote"></div>
  </div>
  <script>//<![CDATA[

(function() {
  var choices = [
    'red',
    'green',
    'blue',
    'yellow',
    'black',
  ];

  var initialSelection = [
    [2, 1, 0],
    [3],
    [],
    [4],
  ];

  new Vote(vote, choices, initialSelection);
})();

  //]]></script>
</body>
</html>`,
				pathSortableJS,
				pathVoteJS,
				pathVoteCSS,
			)
		}
	}
}
