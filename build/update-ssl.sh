#!/bin/bash -e 
#
# Script for updating ca-certificates.crt

prgdir=$(cd $(dirname $0) ; pwd)

echo "Updating certificates"
sudo update-ca-certificates
cp /etc/ssl/certs/ca-certificates.crt $prgdir
