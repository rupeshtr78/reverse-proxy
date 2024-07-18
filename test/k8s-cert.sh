#!/bin/bash
# Get k8s ca cert
kubectl get configmap -n kube-system extension-apiserver-authentication -o jsonpath='{.data.client-ca-file}' > ca.crt