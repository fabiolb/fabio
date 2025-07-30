package ui

import (
	"html/template"
	"net/http"

	"github.com/fabiolb/fabio/config"
)

// RoutesHandler provides the UI for managing the routing table.
type RoutesHandler struct {
	Color        string
	Title        string
	Version      string
	RoutingTable config.RoutingTable
}

func (h *RoutesHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	tmplRoutes.ExecuteTemplate(w, "routes", h)
}

var tmplRoutes = template.Must(template.New("routes").Parse( // language=HTML
	`<!doctype html>
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

		.tooltip { position: relative; display: inline-block; }
		.tooltip .tooltiptext {
			visibility: hidden;
			width: 250px;
			background-color: #000000;
			color: #ffffff;
			text-align: center;
			margin: 30px 0px 0px 30px;
			padding: 5px 0;
			border-radius: 6px;
			position: absolute;
			z-index: 1000;
		}
		.tooltip:hover .tooltiptext { visibility: visible; }
	</style>
</head>
<body>

<ul id="overrides" class="dropdown-content"></ul>

<nav class="top-nav {{.Color}}">

	<div class="container">
		<div class="nav-wrapper">
			<a href="/" class="brand-logo"><img alt="Fabio Logo" style="margin: 15px 0" class="logo" src="/assets/logo.bw.svg"> {{if .Title}} - {{.Title}}{{end}}</a>
			<ul id="nav-mobile" class="right hide-on-med-and-down">
				<li><a class="dropdown-trigger dropdown-button" href="#" data-target="overrides">Overrides<i class="material-icons right">arrow_drop_down</i></a></li>
				<li><a href="https://github.com/fabiolb/fabio/blob/master/CHANGELOG.md">{{.Version}}</a></li>
				<li><a href="https://github.com/fabiolb/fabio">Github</a></li>
				<li><a href="https://fabiolb.net">Fabiolb.net</a></li>
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

	$('.dropdown-trigger').dropdown();
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

		console.log(routes);

		if (routes != null) {
			for (let i=0; i < routes.length; i++) {
				const r = routes[i];
				
				let $tr = $('<tr />');
				if (/^https?:\/+/i.exec(r.src) != null) { 
					$tr = $('<tr />').attr('style', 'background-color: #ff1a1a;');
					$tr.append($('<td />').addClass('tooltip').append($('<span class="tooltiptext"></span>').text("Route Source cannot start with the protocol or scheme (e.g. - 'http' and 'https' are invalid to have listed in the route source)")).append($('<span class="valign-wrapper" />').append(i+1).append($('<i class="material-icons">error_outline</i>'))));
				} else {
					$tr.append($('<td />').text(i+1));
				}
				$tr.append($('<td />').text(r.service));

				if ({{.RoutingTable.Source.LinkEnabled}} == true && /^https?:\/+/i.exec(r.dst) != null && /^https?:\/+/i.exec(r.src) == null) {
					const hrefScheme = ({{.RoutingTable.Source.Scheme}} != '' ? {{.RoutingTable.Source.Scheme}} + ':' : window.location.protocol) + '//';
					const hrefHost = ({{.RoutingTable.Source.Host}} != '' ? {{.RoutingTable.Source.Host}} : window.location.hostname);
					const hrefPort = (/:/gi.exec(r.src) != null ? /:[0-9]*\/?/gi.exec(r.src)[0] : '{{if .RoutingTable.Source.Port}}:{{.RoutingTable.Source.Port}}{{end}}');
					const hrefStr = (r.src.startsWith('/') ? hrefScheme + hrefHost + hrefPort : '{{if .RoutingTable.Source.Scheme}}{{.RoutingTable.Source.Scheme}}:{{end}}//') + r.src;
					$tr.append($('<td />').append($('<a />').attr('href', hrefStr){{if .RoutingTable.Source.NewTab}}.attr('target', '_blank'){{end}}.text(r.src)));
				} else {
					$tr.append($('<td />').text(r.src));
				}

				$tr.append($('<td />').append($('<a />').attr('href', r.dst).text(r.dst)));
				$tr.append($('<td />').text(r.opts));
				$tr.append($('<td />').text((r.weight * 100).toFixed(2) + '%'));

				$tr.appendTo($tbody);
			}
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
			if (val == "") {
				val = "default";
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
