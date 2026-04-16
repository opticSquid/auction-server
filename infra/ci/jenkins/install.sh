#!/bin/bash
helm repo add jenkins https://charts.jenkins.io
helm repo update
helm install jenkins jenkins/jenkins --version 5.9.14 -f ./custom-config.yaml --namespace ci --create-namespace
echo "Waiting 2 mins for Jenkins to start"
sleep 120
echo "Username: admin"
echo "Password: $(kubectl exec --namespace ci -it svc/jenkins -c jenkins -- /bin/cat /run/secrets/additional/chart-admin-password && echo)"
kubectl apply -f jenkins-expose-service.yaml
