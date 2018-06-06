FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
ADD fabio.properties /etc/fabio/fabio.properties
ADD fabio /
EXPOSE 9998 9999
CMD ["/fabio", "-cfg", "/etc/fabio/fabio.properties"]
