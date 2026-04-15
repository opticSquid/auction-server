#!/bin/bash
kind create cluster --config cluster-config.yaml

echo "Cluster created successfully"
echo "Installing Infrastructure..."
echo "[1/2] Installing metrics-server"
./metrics-server/install.sh
echo "[2/2] Installing Jenkins..."
cd ../ci/jenkins/
./install.sh
cd ../../cluster
