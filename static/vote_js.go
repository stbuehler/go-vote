package static

var vote_js = StaticContent{
	Hash:        true,
	FileName:    "vote-##.js",
	ContentType: "application/javascript",
	Body: []byte(`
function Vote(node, choices, initialSelection) {
  var self = this;
  this.node = node;
  this.choices = choices;
  this.selection = initialSelection;
  this.sortables = [];
  this.separators = [];
  this.sorting = {
    group: "group-for-" + this.node.id,
    onAdd: function(evt) {
      var ndx = +evt.to.getAttribute("data-index");
      // handle add
      if (!evt.to.hasAttribute("data-separator")) {
        // not a separator
        self.selection[ndx] = self.sortables[ndx].toArray().map(Number);
      } else {
        // convert separator into real list
        evt.to.removeAttribute("data-separator");
        evt.to.className = "";
        self.sortables.splice(ndx, 0, self.separators[ndx]);
        self.selection.splice(ndx, 0, self.sortables[ndx].toArray().map(Number));
        self.separators.splice(ndx, 1);
        self._addSeparator(ndx, evt.to);
        self._addSeparator(ndx+1, evt.to.nextSibling);
        for (ndx++; ndx < self.sortables.length; ndx++) {
          self.sortables[ndx].el.setAttribute("data-index", ndx);
        }
      }
      // handle remove
      ndx = +evt.from.getAttribute("data-index");
      var sel = self.selection[ndx] = self.sortables[ndx].toArray().map(Number);
      if (0 === sel.length) {
        // remove now empty list and following separator
        self.node.removeChild(self.sortables[ndx].el);
        self.node.removeChild(self.separators[ndx+1].el);
        self.selection.splice(ndx, 1);
        self.sortables.splice(ndx, 1);
        self.separators.splice(ndx+1, 1);
        // fix index
        for (; ndx < self.sortables.length; ndx++) {
          self.sortables[ndx].el.setAttribute("data-index", ndx);
          self.separators[ndx+1].el.setAttribute("data-index", ndx+1);
        }
      }
      console.log("Current vote: " + JSON.stringify(self.selection));
    },
  };
  this.redraw();
}

Vote.prototype._appendList = function(list) {
  var ndx = this.sortables.length;
  list.setAttribute("data-index", ndx);
  this.sortables[ndx] = Sortable.create(list, this.sorting);
  this.node.appendChild(list);
};

Vote.prototype._addSeparator = function(ndx, before) {
  var list = document.createElement('ul');
  list.setAttribute("data-index", ndx);
  list.setAttribute("data-separator", "1");
  list.className = "separator";
  this.separators.splice(ndx, 0, Sortable.create(list, this.sorting));
  this.node.insertBefore(list, before);
  for (ndx++; ndx < this.separators.length; ndx++) {
    this.separators[ndx].el.setAttribute("data-index", ndx);
  }
};

Vote.prototype.clear = function() {
  this.node.innertHtml = "";
  this.sortables = [];
  this.separators = [];
};

Vote.prototype.redraw = function() {
  var i, j, sel, list, elem;
  this.clear();
  for (i = 0; i < this.selection.length; i++) {
    if (0 === this.selection[i].length) {
      this.selection.splice(i--, 1);
    }
  }
  this._addSeparator(0);
  for (i = 0; i < this.selection.length; i++) {
    sel = this.selection[i].sort();
    list = document.createElement('ul');
    for (j = 0; j < sel.length; j++) {
      elem = document.createElement('li');
      elem.setAttribute("data-id", sel[j]);
      elem.innerText = this.choices[sel[j]];
      list.appendChild(elem);
    }
    this._appendList(list);
    this._addSeparator(i+1);
  }
};

Vote.prototype.submit = function(prefix, elId, voter, onfinished) {
  var xhr = new XMLHttpRequest();
  xhr.open('POST', prefix + "/vote?election=" + elId, true);
  xhr.onreadystatechange = function() {
    if (xhr.readyState != 4) return; // not done
    if (onfinished) onfinished();
  };
  xhr.send(JSON.stringify({
    auth: { name: voter },
    rankgroups: this.selection,
  }));
}
`),
}
