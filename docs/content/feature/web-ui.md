---
title: "Web UI"
sincd: "1.0"
---

fabio supports a Web UI to examine the current routing table and manage the
manual overrides. By default it listens on `http://0.0.0.0:9998/` which can be
changed with the `ui.addr` option. The `ui.title` and `ui.color` options allow
customization of the title and the color of the header bar.

The `ui.path` option configures a base path for the UI and API, allowing fabio
to be served behind a reverse proxy at a sub-path (e.g. `ui.path = /fabio`).
