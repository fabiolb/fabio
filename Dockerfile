FROM scratch
ADD build/ca-certificates.crt /etc/ssl/certs/
ADD fabio.properties /etc/fabio/fabio.properties
ADD fabio /
CMD ["/fabio", "-cfg", "/etc/fabio/fabio.properties"]
