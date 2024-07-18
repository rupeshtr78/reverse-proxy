#!/bin/bash

# Variables
CERT_NAME="k8s-p930s-cert"
SERVICE_ACCOUNT="k8s-proxy"
NAMESPACE="default"

kubectl create serviceaccount k8s-proxy

# Generate a new private key if necessary
openssl genrsa -out ${CERT_NAME}.key 2048

# Create a certificate signing request (CSR)
openssl req -new -key ${CERT_NAME}.key -subj "/CN=${SERVICE_ACCOUNT}.${NAMESPACE}.svc" -out ${CERT_NAME}.csr

# Create a CSR object in Kubernetes
kubectl apply -f -

apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${CERT_NAME}
spec:
  groups:
  - system:authenticated
  request: $(cat ${CERT_NAME}.csr | base64 | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth


# Use Kubernetes' CA to sign the CSR and get back a certificate
kubectl certificate approve ${CERT_NAME}

# Get the signed certificate
kubectl get csr ${CERT_NAME} -o jsonpath='{.status.certificate}' | base64 --decode > ${CERT_NAME}.crt

# Create a secret from the certificate in Kubernetes
kubectl create secret tls ${CERT_NAME} --key=${CERT_NAME}.key --cert=${CERT_NAME}.crt

# Patch the secret to the service account
kubectl patch serviceaccount ${SERVICE_ACCOUNT} -p "{\"secrets\": [{\"name\": \"${CERT_NAME}\"}]}"