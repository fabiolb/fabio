# fabiolb.net website

This is the source code for the https://fabiolb.net website.

It is built with [Hugo](https://gohugo.io/) and deployed automatically
to [Netlify](https://netlify.com/).

The theme is [TheDocs](http://thetheme.io/thedocs/) from [TheTheme.io](http://thetheme.io/).

This isn't a free theme and I've paid for it. Please don't just take it. Pay the designers.
It isn't expensive.

To render the page locally run the following command in the `fabio` root directory:

    $ hugo serve -s docs --disableFastRender

To view the site open http://localhost:1313/ in your browser.

## Content organization

The content is organized in sections which each live in their own sub-directory.
The menu is generated from the sections and pages within sections. Ordering is done
with the `weight` attribute in the front-matter (aka. top of the `_index.md` files).
