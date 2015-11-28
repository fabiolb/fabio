package ui

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/eBay/fabio/route"
)

func handleRoute(w http.ResponseWriter, r *http.Request) {
	var cfg [][]string
	for _, s := range route.GetTable().Config(true) {
		p := strings.Split(s, "tags")
		if len(p) == 1 {
			cfg = append(cfg, []string{s, ""})
		} else {
			cfg = append(cfg, []string{strings.TrimSpace(p[0]), "tags" + p[1]})
		}
	}

	data := struct {
		Config    [][]string
		ConfigURL string
	}{
		cfg,
		configURL,
	}
	tmplTable.ExecuteTemplate(w, "table", data)
}

var tmplTable = template.Must(template.New("table").Parse(htmlTable))

var htmlTable = `
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>fabio routing table</title>
  <script src="https://code.jquery.com/jquery-1.10.2.js"></script>
  <style type="text/css">
    *, html, body {font-family: sans-serif;}
  	table, tr, td {text-align: left;}
  	td, input {font-family: monospace; font-size: 12px;}
  </style>
</head>
<body>
<h2>./fabio - Routing Table</h2>

<p>Filter routes: <input type="text" id="filter"></p>

<table>
<tbody>
{{range $i, $v := .Config}}
	<tr><td class="route">{{index $v 0}}</td><td>{{index $v 1}}</td></tr>
{{end}}
</tbody>
</table>

<p><a href="{{.ConfigURL}}" target="_new">Edit config</a></p>

<script>

$(function(){
	var $filter = $('#filter');
	var params={};window.location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(str,key,value){params[key] = value;});
	if (params.filter) {
		$filter.val(params.filter);
		$("td.route:not(:contains('"+params.filter+"'))").each(function() {
			$(this).parent("tr").hide();
		});
	}
	$filter.focus();
	$filter.keyup(function() {
		var url = window.location.href.split('?')[0];
		if ($filter.val() != '') {
			url +=  "?filter=" +$filter.val();
		}
		window.location = url;
	})
})
</script>

</body>
</html>
`
