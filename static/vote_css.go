package static

var vote_css = StaticContent{
	Hash:        true,
	FileName:    "vote-##.css",
	ContentType: "text/css; charset=utf-8",
	Body: []byte(`

div.block {
  max-width: 400px;
  margin: 10px;
  display: inline-block;
}

#vote ul {
  list-style: none;
  padding: 10px;
  margin: 10px 0;
  border-radius: 5px;
  /*border: 1px solid red;*/
  background: #aaa;
}

#vote ul.separator {
  background: #eee;
  min-height: 30px;
}

#vote ul::after {
  clear: both;
  content: '';
  display: block;
}

#vote li {
  background: #5F9EDF;
  color: white;
  text-align: center;
  float: left;
  margin: 5px;
  padding: 10px 15px;
  cursor: move;
  border-radius: 5px;
  min-width: 50px;
}

#vote li.sortable-ghost {
  opacity: .3;
  background: #f60;
}

table.winning {
  border-collapse: collapse;
}
table.winning td, table.winning th {
  padding: 3px 5px;
  border: 1px solid grey;
  margin: 0;
}
table.winning td {
  text-align: right;
}
table.winning td.win {
  background: green;
}
table.winning td.loose {
  background: red;
}
`),
}
