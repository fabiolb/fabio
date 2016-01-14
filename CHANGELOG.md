## Changelog

### [v1.0.9] - Unreleased

 * [Issue #37](https://github.com/eBay/fabio/issues/37): Add support for consul ACL token to demo server

### [v1.0.8](https://github.com/eBay/fabio/releases/tag/v1.0.8) - 14 Jan 2015

 * Upgrade to Go 1.5.3
 * [Issue #29](https://github.com/eBay/fabio/issues/29): Include service with check ids other than 'service:*'
 * [Issue #30](https://github.com/eBay/fabio/issues/30): Register fabio with local ip address as fallback
 * [Issue #36](https://github.com/eBay/fabio/issues/36): Add support for consul ACL token

### [v1.0.7](https://github.com/eBay/fabio/releases/tag/v1.0.7) - 13 Dec 2015

 * [Issue #22](https://github.com/eBay/fabio/issues/22): fabio route not removed after consul deregister
 * [Issue #23](https://github.com/eBay/fabio/issues/23): routes not removed when passing empty string
 * [Issue #26](https://github.com/eBay/fabio/issues/26): Detect when consul agent is down
 * Allow to override title and color UI

### [v1.0.6](https://github.com/eBay/fabio/releases/tag/v1.0.6) - 01 Dec 2015

 * [Issue #9](https://github.com/eBay/fabio/issues/9): Enabled raw websocket proxy by default
 * [Issue #15](https://github.com/eBay/fabio/issues/15): Traffic shaping now matches on service
 * [Issue #16](https://github.com/eBay/fabio/issues/16): Improved Web UI with better filtering
 * [Issue #18](https://github.com/eBay/fabio/issues/18): Manage manual overrides via ui

### [v1.0.5](https://github.com/eBay/fabio/releases/tag/v1.0.5) - 11 Nov 2015

 * [Issue #9](https://github.com/eBay/fabio/issues/9): Add experimental support for web sockets
 * [Issue #10](https://github.com/eBay/fabio/issues/10): Add support for `Forwarded` and `X-Forwarded-For` header
 * Add `proxy.localip` to set proxy ip address for headers

### [v1.0.4](https://github.com/eBay/fabio/releases/tag/v1.0.4) - 03 Nov 2015

 * [Issue #8](https://github.com/eBay/fabio/issues/8): Add support for SSL client certificate authentication

### [v1.0.3](https://github.com/eBay/fabio/releases/tag/v1.0.3) - 25 Oct 2015

 * Add Docker support and official Docker image `magiconair/fabio`

### [v1.0.2](https://github.com/eBay/fabio/releases/tag/v1.0.2) - 23 Oct 2015

 * [Pull Request #3](https://github.com/eBay/fabio/pull/3): Honor consul.url and consul.addr from config file (@jeinwag)

### [v1.0.1](https://github.com/eBay/fabio/releases/tag/v1.0.1) - 21 Oct 2015

 * Honor maintenance mode for both services and nodes

### [v1.0.0](https://github.com/eBay/fabio/releases/tag/v1.0.0) - 16 Oct 2015

 * Initial open-source release

