#!/bin/bash
 
#  The  SSL_CERT_DIR  environment variable is used to specify the directory where the certificates are stored. 
#  The  go run main.go  command is used to run the Go program. 
#  The  main.go  file is the entry point of the Go program. 
#  The  main.go  file contains the following code:

SSL_CERT_DIR=/home/rupesh/aqrtr/security/ssl/ca/keystore/ca.crt go run ./cmd/main.go

# Test the Go program
curl -H "Authorization: Bearer $TOKEN" https://10.0.0.213:6445/api/v1/namespaces/default/pods