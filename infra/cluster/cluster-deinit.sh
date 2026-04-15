#!/bin/bash

echo "Will DELETE the entire Cluster are you sure [y/N]?"
read -r response

if [[ $response =~ ^[Yy]$ ]]; then
    helm uninstall jenkins --namespace ci
    helm uninstall metrics-server --namespace kube-system
    kind delete cluster
fi
