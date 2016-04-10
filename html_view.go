package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JsonMustEncode(v interface{}) []byte {
	if s, err := json.Marshal(v); nil != err {
		panic(err)
	} else {
		return s
	}
}

func JsonMustEncodeString(v interface{}) string {
	return string(JsonMustEncode(v))
}

func (e *Election) ViewHTML(baseURL string, elId string, w http.ResponseWriter) {
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Vote</title>

  <script src="%[1]sSortable-1.4.2.min.js"></script>
  <script src="%[1]svote-0.1.js"></script>
  <link rel="stylesheet" href="%[1]svote-0.1.css">
</head>
<body style="text-align: center;">
  <div style="display: inline-block; text-align: left;">
    <div class="vote-block" style="max-width: 400px; margin: 10px; display: inline-block;">
      <h2>Rank according to your preferences</h2>
      <p>Choices in the same block have equal preference. Choices in blocks at the top are preferred over choices in lower blocks.</p>
      <div id="vote"></div>
      <p><label>Name: <input id="voter" type="text" size="30"></input></label></p>
      <p><button id="submit-vote">Submit</button></p>
    </div>
    <div class="vote-block" style="max-width: 400px; margin: 10px; display: inline-block;">
      <h2>Result</h2>
      <p><button id="submit-result">Reload</button></p>
      <div id="result"></div>
    </div>
  </div>
  <script>//<![CDATA[

(function() {
  var choices = %[3]s;

  function make_winning_table(numbers) {
    var i, j, table, row, cell, diff;

    table = document.createElement("table");
    table.className = "winning";

    row = document.createElement("tr");
    row.appendChild(document.createElement("th"));
    for (j = 0; j < choices.length; j++) {
      cell = document.createElement("th");
      cell.innerText = choices[j];
      row.appendChild(cell);
    }
    table.appendChild(row);

    for (i = 0; i < choices.length; i++) {
      row = document.createElement("tr");

      cell = document.createElement("th");
      cell.innerText = choices[i];
      row.appendChild(cell);

      for (j = 0; j < choices.length; j++) {
        cell = document.createElement("td");
        if (i == j) {
          cell.className = "self";
        } else {
          diff = numbers[i][j] - numbers[j][i];
          cell.innerText = numbers[i][j];
          cell.className = (diff > 0) ? "win" : (diff < 0) ? "loose" : "";
        }
        row.appendChild(cell);
      }
      table.appendChild(row);
    }

    return table;
  }

  var r = document.getElementById('result')
  function show_result(result) {
    var p;
    r.innerText = ""; //JSON.stringify(result);

    p = document.createElement("p");
    if (0 === result.winner || result.winner) {
      p.innerText = "The winner is: " + choices[result.winner];
    } else {
      p.innerText = "There is no winner";
    }
    r.appendChild(p);

    p = document.createElement("p");
    p.innerText = "How often row wins over column:";
    r.appendChild(p);
    r.appendChild(make_winning_table(result.preferences));

    if (result.paths) {
      p = document.createElement("p");
      p.innerText = "Strengths of strongest paths for Schwarz method:";
      r.appendChild(p);
      r.appendChild(make_winning_table(result.paths));
    }
  }

  function load_result() {
    var xhr = new XMLHttpRequest();
    xhr.responseType = "json";
    xhr.open('GET', "%[1]se/result/" + %[2]s, true);
    xhr.onreadystatechange = function() {
      if (xhr.readyState != 4) return; // not done
      show_result(xhr.response);
    };
    xhr.send();
  }

  v = new Vote(document.getElementById('vote'), choices, %[4]s);
  document.getElementById('submit-vote').onclick = function() {
    v.submit("%[1]s", %[2]s, document.getElementById('voter').value, load_result);
  };

  document.getElementById('submit-result').onclick = load_result;
  load_result();
})();

  //]]></script>
</body>
</html>`, baseURL, JsonMustEncodeString(elId), JsonMustEncodeString(e.Choices), JsonMustEncodeString(e.SanitizeVote(Vote{})))
}
