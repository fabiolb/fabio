package ui

import (
	"html/template"
	"net/http"
)

// HandleManual provides the UI for the manual overrides.
func HandleManual(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Color   string
		Title   string
		Version string
	}{Color, Title, Version}
	tmplManual.ExecuteTemplate(w, "manual", data)
}

var tmplManual = template.Must(template.New("manual").Parse(`
<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>./fabio{{if .Title}} - {{.Title}}{{end}}</title>
	<script type="text/javascript" src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.3/css/materialize.min.css">
	<script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/0.97.3/js/materialize.min.js"></script>
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>

	<style type="text/css">
	</style>
</head>
<body>

<nav class="top-nav {{.Color}}">

	<div class="container">
		<div class="nav-wrapper">
			<a href="/" class="brand-logo">./fabio{{if .Title}} - {{.Title}}{{end}}</a>
			<ul id="nav-mobile" class="right hide-on-med-and-down">
				<li><a href="/routes">Routes</a></li>
				<li><a href="https://github.com/eBay/fabio/blob/master/CHANGELOG.md">{{.Version}}</a></li>
				<li><a href="https://github.com/eBay/fabio">Github</a></li>
			</ul>
		</div>
	</div>

</nav>

<div class="container">

	<div class="section">
		<h5>Manual Overrides</h5>

		<div class="row">
			<form class="col s12">
				<input type="hidden" name="version">
				<div class="row">
					<div class="input-field col s12">
						<textarea id="textarea1" class="materialize-textarea"></textarea>
						<label for="textarea1"></label>
					</div>
				</div>
			</form>
			<button class="btn waves-effect waves-light" name="save">Save</button>
			<button class="btn waves-effect waves-light" name="help">Help</button>
		</div>

		<div class="row">
			<pre class="help hide">
route add &lt;svc&gt; &lt;src&gt; &lt;dst&gt; weight &lt;w&gt; tags "&lt;t1&gt;,&lt;t2&gt;,..."
  - Add route for service svc from src to dst and assign weight and tags

route add &lt;svc&gt; &lt;src&gt; &lt;dst&gt; weight &lt;w&gt;
  - Add route for service svc from src to dst and assign weight

route add &lt;svc&gt; &lt;src&gt; &lt;dst&gt; tags "&lt;t1&gt;,&lt;t2&gt;,..."
  - Add route for service svc from src to dst and assign tags

route add &lt;svc&gt; &lt;src&gt; &lt;dst&gt;
  - Add route for service svc from src to dst

route del &lt;svc&gt; &lt;src&gt; &lt;dst&gt;
  - Remove route matching svc, src and dst

route del &lt;svc&gt; &lt;src&gt;
  - Remove all routes of services matching svc and src

route del &lt;svc&gt;
  - Remove all routes of service matching svc

 route del &lt;svc&gt; tags "&lt;t1&gt;lt;t2&gt;..."
   - Remove all routes of service matching svc and tags

 route del tags "&lt;t1&gt;lt;t2&gt;..."
   - Remove all routes matching tags

route weight &lt;svc&gt; &lt;src&gt; weight &lt;w&gt; tags "&lt;t1&gt;,&lt;t2&gt;,..."
  - Route w% of traffic to all services matching svc, src and tags

route weight &lt;src&gt; weight &lt;w&gt; tags "&lt;t1&gt;,&lt;t2&gt;,..."
  - Route w% of traffic to all services matching src and tags

route weight &lt;svc&gt; &lt;src&gt; weight &lt;w&gt;
  - Route w% of traffic to all services matching svc and src

route weight service host/path weight w tags "tag1,tag2"
  - Route w% of traffic to all services matching service, host/path and tags

    w is a float &gt; 0 describing a percentage, e.g. 0.5 == 50%
    w &lt;= 0: means no fixed weighting. Traffic is evenly distributed
    w &gt; 0: route will receive n% of traffic. If sum(w) &gt; 1 then w is normalized.
    sum(w) &gt;= 1: only matching services will receive traffic

    Note that the total sum of traffic sent to all matching routes is w%.
			</pre>
		</div>
	</div>

</div>

<script>
$(function(){
	var params={};window.location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(str,key,value){params[key] = value;});

	$.get("/api/manual", function(data) {
		$("input[name=version]").val(data.version);
		$("textarea>label").val("Version " + data.version);
		$("#textarea1").val(data.value);
		$("#textarea1").trigger('autoresize');
	});

	$("button[name=help]").click(function() {
		$("pre.help").toggleClass("hide");
	});

	$("button[name=save]").click(function() {
		var data = {
			value   : $("#textarea1").val(),
			version : $("input[name=version]").val()
		}
		$.ajax('/api/manual', {
			type: 'PUT',
			data: JSON.stringify(data),
			contentType: 'application/json',
			statusCode: {
				400: function(jqXHR, textStatus, err) { alert(err); },
				409: function(jqXHR, textStatus, err) { alert(err); },
				500: function(jqXHR, textStatus, err) { alert(err); }
			},
			success: function() {
				window.location.reload();
			}
		});
	});
})
</script>

</body>
</html>
`))
