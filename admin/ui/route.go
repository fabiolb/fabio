package ui

import (
	"html/template"
	"net/http"
)

// RoutesHandler provides the UI for managing the routing table.
type RoutesHandler struct {
	Color   string
	Title   string
	Version string
}

func (h *RoutesHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	tmplRoutes.ExecuteTemplate(w, "routes", h)
}

var tmplRoutes = template.Must(template.New("routes").Parse( // language=HTML
	`
<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>fabio{{if .Title}} - {{.Title}}{{end}}</title>
	<script type="text/javascript" src="/assets/code.jquery.com/jquery-3.6.0.min.js"></script>
    <link href="/assets/fonts/material-icons.css" rel="stylesheet">
    <link rel="stylesheet" href="/assets/cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css">
    <script src="/assets/cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js"></script>
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
		<img alt="Fabio Logo" class="logo" src="/assets/logo.svg">
	</div>

</div>

<script>
$(function(){
	let params={};window.location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(str,key,value){params[key] = value;});

	function renderRoutes(routes) {
		const $table = $('table.routes');

		let thead = '<thead><tr>';
		thead += '<th>#</th>';
		thead += '<th>Service</th>';
		thead += '<th>Source</th>';
		thead += '<th>Dest</th>';
		thead += '<th>Options</th>';
		thead += '<th>Weight</th>';
		thead += '</tr></thead>';

		let $tbody = $('<tbody />');

		for (let i=0; i < routes.length; i++) {
			const r = routes[i];

			const $tr = $('<tr />')

			$tr.append($('<td />').text(i+1));
			$tr.append($('<td />').text(r.service));
			$tr.append($('<td />').text(r.src));
			$tr.append($('<td />').append($('<a />').attr('href', r.dst).text(r.dst)));
			$tr.append($('<td />').text(r.opts));
			$tr.append($('<td />').text((r.weight * 100).toFixed(2) + '%'));

			$tr.appendTo($tbody);
		}

		$table.empty().
			append($(thead)).
			append($tbody);
	}

	let $filter = $('#filter');
	function doFilter(v) {
		$("tr").show();
		if (!v) return;
		let words = v.split(' ');
		for (let i=0; i < words.length; i++) {
			let w = words[i].trim();
			if (w === "") continue;
			$("tbody tr:not(:contains('"+w+"'))").hide();
		}
	}

	$filter.focus();
	$filter.keyup(function() {
		const v = $filter.val();
		window.history.pushState(null, null, "?filter=" +v);
		doFilter(v);
	});

	$.get("/api/routes", function(data) {
		renderRoutes(data);
		if (!params.filter) return;
		const v = decodeURIComponent(params.filter);
		$filter.val(v);
		doFilter(v);
	});

	$.get('/api/paths', function(data) {
		const d = $("#overrides");
		$.each(data, function(idx, val) {
			let path = val;
			if (val === "") {
				val = "default"
			}
			d.append(
				$('<li />').append(
					$('<a />').attr('href', '/manual'+path).text(val)
				)
			);
		});
	});
});
</script>

</body>
</html>
`))
