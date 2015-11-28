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
	<title>./fabio</title>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.3/css/materialize.min.css">
	<script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.3/js/materialize.min.js"></script>
	<script type="text/javascript" src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>

	<style type="text/css">
		td.tags { display: none; }

		/*
		 * -- DESKTOP (AND UP) MEDIA QUERIES --
		 * On desktops and other large devices, we want to over-ride some
		 * of the mobile and tablet styles.
		 */
		@media (min-width: 78em) {
			td.tags{ display: table-cell; }
		}
	</style>
</head>
<body>

<nav class="top-nav light-green">

	<div class="container">
		<div class="nav-wrapper">
			<a href="https://github.com/eBay/fabio" class="brand-logo">./fabio</a>
			<ul id="nav-mobile" class="right hide-on-med-and-down">
				<li><a href="{{.ConfigURL}}">consul KV</a></li>
				<li><a href="https://github.com/eBay/fabio/blob/master/CHANGELOG.md">{{.Version}}</a></li>
			</ul>
		</div>
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

	function doFilter(v) {
		$("tr").show();
		$filter.val(v);
		if (!v || v == "") return;
		$("td.route:not(:contains('"+v+"'))").each(function() {
			$(this).parent("tr").hide();
		});
	}

	var params={};window.location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(str,key,value){params[key] = value;});
	doFilter(params.filter);

	$filter.focus();
	$filter.keyup(function() {
		var v = $filter.val();
		window.history.pushState(null, null, "?filter=" +v);
		doFilter(v);
	});
})
</script>

</body>
</html>
`
