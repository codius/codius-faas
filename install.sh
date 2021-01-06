#!/bin/bash

set -e

kubectl set env -n openfaas-fn deploy/system-github-event validate_customers=false

kubectl set env -n openfaas deploy/edge-router auth_url=http://edge-auth.openfaas:8080

kubectl apply -f https://raw.githubusercontent.com/wilsonianb/faas-netes/master/artifacts/crds/openfaas.com_profiles.yaml
kubectl patch -n openfaas deploy/gateway -p '
spec:
  template:
    spec:
      containers:
      - name: faas-netes
        image: wilsonianbcoil/faas-netes
        imagePullPolicy: Always
'

kubectl patch serviceaccount -n openfaas-fn default -p '{"automountServiceAccountToken": false}'

export PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath='{.data.basic-auth-password}' | base64 --decode | openssl passwd -crypt -stdin)

kubectl create secret generic ingress-basic-auth \
 --from-literal=admin=$PASSWORD --namespace openfaas \
 --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -k config/default

kubectl patch -n openfaas-fn deploy/buildshiprun -p '
spec:
  template:
    spec:
      containers:
      - name: buildshiprun
        image: ghcr.io/codius/ofc-buildshiprun:latest
        imagePullPolicy: Always
        env:
        - name: profile
          value: ofc-workload
'

kubectl patch -n openfaas-fn deploy/system-dashboard -p '
spec:
  template:
    spec:
      containers:
      - name: system-dashboard
        image: ghcr.io/codius/of-cloud-dashboard:latest
        imagePullPolicy: Always
        env:
        - name: receipts_url
          valueFrom:
            configMapKeyRef:
              key: receipts_url
              name: system-dashboard
        - name: payment_pointer
          valueFrom:
            configMapKeyRef:
              key: payment_pointer
              name: system-dashboard
'

export PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode; echo)

export $(grep -v '^#' config.env | xargs)
export OPENFAAS_URL=https://gateway.${root_domain}
echo -n $PASSWORD | faas-cli login --username admin --password-stdin
faas-cli template store pull golang-middleware
faas-cli template store pull node12
TAG=latest faas-cli deploy -f ./stack.yml
