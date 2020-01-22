# fabiolb.net website

This is the source code for the https://fabiolb.net website.

It is built with [Hugo](https://gohugo.io/) and automatically deployed
via [Bamboo](https://www.atlassian.com/software/bamboo) and
[Nomad](https://www.nomadproject.io/) to [ENA](https://github.com/myENA)'s Docker infrastructure
and exposed through [Consul](https://consul.io/) and [Fabio](https://fabiolb.net/).

The theme is [TheDocs](http://thetheme.io/thedocs/) from [TheTheme.io](http://thetheme.io/).

This isn't a free theme and I've paid for it. Please don't just take it. Pay the designers.
It isn't expensive.

To render the page locally run the following command in the `fabio` root directory:

    $ hugo serve -s docs --disableFastRender

To view the site open http://localhost:1313/ in your browser.

