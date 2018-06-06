package ui

import (
	"html/template"
	"net/http"
)

// RoutesHandler provides the UI for managing the routing table.
type RoutesHandler struct {
	Color, Title, Version string
}

func (h *RoutesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmplRoutes.ExecuteTemplate(w, "routes", h)
}

var tmplRoutes = template.Must(template.New("routes").Parse(`
<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>fabio{{if .Title}} - {{.Title}}{{end}}</title>
	<script type="text/javascript" src="https://code.jquery.com/jquery-3.3.1.min.js"></script>
    <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.100.2/css/materialize.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.100.2/js/materialize.min.js"></script>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>

	<style type="text/css">
		td.tags { display: none; }
		.footer { padding-top: 10px; }
		.logo { height: 32px; margin: 0 auto; display: block; }

		@media (min-width: 78em) {
			td.tags{ display: table-cell; }
		}
	</style>
</head>
<body>

<ul id="overrides" class="dropdown-content"></ul>

<nav class="top-nav {{.Color}}">

	<div class="container">
		<div class="nav-wrapper">
			<a href="/" class="brand-logo">fabio{{if .Title}} - {{.Title}}{{end}}</a>
			<ul id="nav-mobile" class="right hide-on-med-and-down">
                <li><a class="dropdown-button" href="#!" data-activates="overrides">Overrides<i class="material-icons right">arrow_drop_down</i></a></li>
				<li><a href="https://github.com/fabiolb/fabio/blob/master/CHANGELOG.md">{{.Version}}</a></li>
				<li><a href="https://github.com/fabiolb/fabio">Github</a></li>
			</ul>
		</div>
	</div>

</nav>

<div class="container">

	<div class="section">
		<h5>Routing Table</h5>
		<p><input type="text" id="filter" placeholder="type to filter routes"></p>
		<table class="routes highlight"></table>
	</div>

	<div class="section footer">
		<img class="logo" src="/logo.svg">
	</div>

</div>

<script>
$(function(){
	var params={};window.location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(str,key,value){params[key] = value;});

	function renderRoutes(routes) {
		var $table = $("table.routes");

		var tbl = '<thead><tr>';
		tbl += '<th>#</th>';
		tbl += '<th>Service</th>';
		tbl += '<th>Source</th>';
		tbl += '<th>Dest</th>';
		tbl += '<th>Options</th>';
		tbl += '<th>Weight</th>';
		tbl += '</tr></thead><tbody>'
		tbl += '<tbody>'
		for (var i=0; i < routes.length; i++) {
			var r = routes[i];
			tbl += '<tr>';
			tbl += '<td>' + (i+1) + '</td>';
			tbl += '<td>' + r.service + '</td>';
			tbl += '<td>' + r.src + '</td>';
			tbl += '<td>' + r.dst + '</td>';
			tbl += '<td>' + r.opts + '</td>';
			tbl += '<td>' + (r.weight * 100).toFixed(2) + '%</td>';
			tbl += '</tr>';
		}
		tbl += '</tbody>';
		$table.html(tbl);
	}

	var $filter = $('#filter');
	function doFilter(v) {
		$("tr").show();
		if (!v) return;
		var words = v.split(' ');
		console.log('words: ', words);
		for (var i=0; i < words.length; i++) {
			var w = words[i].trim();
			if (w == "") continue;
			$("tbody tr:not(:contains('"+w+"'))").hide();
		}
	}

	$filter.focus();
	$filter.keyup(function() {
		var v = $filter.val();
		window.history.pushState(null, null, "?filter=" +v);
		doFilter(v);
	});

	$.get("/api/routes", function(data) {
		renderRoutes(data);
		if (!params.filter) return;
		var v = decodeURIComponent(params.filter);
		$filter.val(v);
		doFilter(v);
	});

	$.get('/api/paths', function(data) {
		var d = $("#overrides");
		$.each(data, function(idx, val) {
			var path = val;
			if (val == "") {
				val = "default"
			}
            d.append('<li><a href="/manual'+path+'">'+val+'</a></li>');
		});
	});

})
</script>

</body>
</html>
`))
