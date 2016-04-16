package static

var api_js = StaticContent{
	Hash:        true,
	FileName:    "api-##.js",
	ContentType: "application/javascript",
	Body: []byte(`
function setup(prefix, electionName, choices, rankGroups) {
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
    xhr.open('POST', prefix + "/result?election=" + electionName, true);
    xhr.onreadystatechange = function() {
      if (xhr.readyState != 4) return; // not done
      show_result(xhr.response);
    };
    xhr.send(JSON.stringify({
      auth: {},
    }));
  }

  v = new Vote(document.getElementById('vote'), choices, rankGroups);
  document.getElementById('submit-vote').onclick = function() {
    v.submit(prefix, electionName, document.getElementById('voter').value, load_result);
  };

  document.getElementById('submit-result').onclick = load_result;
  load_result();
}
`),
}
