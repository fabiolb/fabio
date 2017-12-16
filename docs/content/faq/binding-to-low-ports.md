---
title: "Binding to Low Ports"
---

If you want to bind fabio to ports below 1024 - so called privileged ports -
without running fabio as `root` you can use an operating system approach as
described below.

These best practices are taken from https://github.com/fabiolb/fabio/issues/195.

#### Linux

Provide `net_bind_service` capability to fabio binary

```
$ setcap 'cap_net_bind_service=+ep' $(which fabio)
```

When using `systemd` you can use the following service definition:

```
$ cat /etc/systemd/system/fabio.service

[Unit]
Description=Fabio proxy
After=syslog.target
After=network.target

[Service]
LimitMEMLOCK=infinity
LimitNOFILE=65535
Type=simple
# unprivileged uid and gid
User=fabio_user
Group=fabio_group
WorkingDirectory=/
ExecStart=/path/to/fabio -cfg /path/to/fabio.conf
Restart=always
# no need that fabio messes with /dev
PrivateDevices=yes
# dedicated /tmp
PrivateTmp=yes
# make /usr, /boot, /etc read only
ProtectSystem=full
# /home is not accessible at all
ProtectHome=yes
# to be able to bind port < 1024
AmbientCapabilities=CAP_NET_BIND_SERVICE
NoNewPrivileges=yes
# only ipv4, ipv6, unix socket and netlink networking is possible
# netlink is necessary so that fabio can list available IPs on startup
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX AF_NETLINK
```

#### Solaris/Illumos/SmartOS

Provide `net_privaddr` privileges to fabio user

```shell
$ /usr/sbin/usermod -K defaultpriv=basic,net_privaddr fabio_user

$ grep fabio_user /etc/user_attr
fabio_user::::type=normal;defaultpriv=basic,net_privaddr
```

#### Provide privilege to fabio process (syntax needs review)

```shell
$ /usr/sbin/ppriv -s PELI+NET_PRIVADDR -e fabio
```

#### OpenBSD/FreeBSD/NetBSD

Use `PF` to forward from low port to high port.

```
/etc/pf.conf

EXT_IF = "eth0"
HTTPS_PORT = 443
HTTPS_PORT_BACKEND = 4343
LOCAL_IP = "127.0.0.1"

...

pass in quick on $EXT_IF inet proto tcp from any to $LOCAL_IP port $HTTPS_PORT rdr-to $LOCAL_IP port $HTTPS_PORT_BACKEND

```

#### FreeBSD: Change the range of reserved ports (this looks dangerous)

```shell
$ sysctl net.inet.ip.portrange.reservedhigh=79

# add to /etc/sysctl.conf to make this permament
```

#### macOS (needs review by SME)

Use `launchd` to launch fabio by creating a service plist and using launchctl to run it:

`$sudo launchctl load -w /path/to/fabio.plist`

Example plist XML (needs reviewing):
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist
  PUBLIC '-//Apple//DTD PLIST 1.0//EN'
  'http://www.apple.com/DTDs/PropertyList-1.0.dtd'>
<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>com.github.fabiolb.fabio</string>
		<key>Program</key>
		<string>fabio</string>
		<key>Sockets</key>
		<dict>
			<key>Listeners</key>
			<dict>
				<key>SockServiceName</key>
				<string>80</string>
				<key>SockType</key>
				<string>stream</string>
				<key>SockFamily</key>
				<string>IPv4</string>
			</dict>
		</dict>
	</dict>
</plist>
```

#### Windows

???
