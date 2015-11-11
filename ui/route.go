package ui

import (
	"html/template"
	"net/http"

	"github.com/eBay/fabio/route"
)

func handleRoute(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Config    []string
		ConfigURL string
	}{
		route.GetTable().Config(true),
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
  	pre, input { font-size: 14px; }
  </style>
</head>
<body>
<h1>Routing Table</h1>

<p>Filter routes: <input type="text" id="filter"></p>

{{range $i, $v := .Config}}
<pre>{{$v}}</pre>
{{end}}

<p><a href="{{.ConfigURL}}" target="_new">Edit config</a></p>

<script>
$filter = $('#filter');
$filter.focus();
$filter.keyup(function() {
	$("pre").show();
	if ($filter.val() == '') {
		return;
	}
	$("pre:not(:contains('"+$filter.val()+"'))").hide();
})
</script>

</body>
</html>
`
