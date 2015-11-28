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
		Version   string
	}{
		cfg,
		configURL,
		version,
	}
	tmplTable.ExecuteTemplate(w, "table", data)
}

func add(x, y int) int {
	return x + y
}

var funcs = template.FuncMap{"add": add}

var tmplTable = template.Must(template.New("table").Funcs(funcs).Parse(htmlTable))

var htmlTable = `
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>fabio routing table</title>
  <script type="text/javascript" src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.3/css/materialize.min.css">
  <script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.3/js/materialize.min.js"></script>

  <style type="text/css">
  </style>
</head>
<body>

<nav>
	<div class="nav-wrapper light-green">
		&nbsp;&nbsp;&nbsp;
		<a href="https://github.com/eBay/fabio" class="brand-logo">./fabio</a>
		<ul id="nav-mobile" class="right hide-on-med-and-down">
			<li><a href="{{.ConfigURL}}">consul KV</a></li>
			<li><a href="https://github.com/eBay/fabio/blob/master/CHANGELOG.md">{{.Version}}</a></li>
		</ul>
	</div>
</nav>

<div class="container">

	<div class="section">
		<h5>Routing Table</h5>

		<p><input type="text" id="filter" placeholder="type to filter routes"></p>

		<table class="highlight">
		<tbody>
		{{range $i, $v := .Config}}<tr>
			<td class="idx">{{add $i 1}}.</td>
			<td class="route">{{index $v 0}}</td>
			<td class="tags">{{index $v 1}}</td>
		</tr>
		{{end}}</tbody>
		</table>
	</div>

</div>

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
