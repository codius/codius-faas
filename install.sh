kubectl set env -n openfaas-fn deploy/system-github-event validate_customers=false

kubectl patch -n openfaas deploy/edge-router -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "edge-router",
          "image": "wilsonianbcoil/edge-router",
          "imagePullPolicy": "Always",
          "env": [{
            "name": "auth_url",
            "value": "http://edge-auth.openfaas:8080"
          }]
        }]
      }
    }
  }
}'

kubectl apply -f https://raw.githubusercontent.com/wilsonianb/faas-netes/master/artifacts/crds/openfaas.com_profiles.yaml
kubectl patch -n openfaas deploy/gateway -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "faas-netes",
          "image": "wilsonianbcoil/faas-netes",
          "imagePullPolicy": "Always"
        }]
      }
    }
  }
}'

kubectl patch serviceaccount -n openfaas-fn default -p '{"automountServiceAccountToken": false}'

kubectl apply -k .

kubectl annotate ingress -n openfaas openfaas-ingress nginx.ingress.kubernetes.io/custom-http-errors=402 nginx.ingress.kubernetes.io/default-backend=svc-402-page

kubectl patch -n openfaas-fn deploy/buildshiprun -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "buildshiprun",
          "image": "wilsonianbcoil/of-buildshiprun",
          "imagePullPolicy": "Always",
          "env": [{
            "name": "profile",
            "value": "ofc-workload"
          }]
        }]
      }
    }
  }
}'

kubectl patch -n openfaas-fn deploy/system-dashboard -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "system-dashboard",
          "image": "wilsonianbcoil/of-cloud-dashboard",
          "imagePullPolicy": "Always",
          "env": [{
            "name": "github_app_url",
            "valueFrom": {
              "configMapKeyRef": {
                "key": "github_app_url",
                "name": "github-app"
              }
            }
          }]
        }]
      }
    }
  }
}'
