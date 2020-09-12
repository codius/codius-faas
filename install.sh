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

export PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath='{.data.basic-auth-password}' | base64 --decode | openssl passwd -crypt -stdin)

kubectl create secret generic ingress-basic-auth \
 --from-literal=admin=$PASSWORD --namespace openfaas \
 --dry-run=client -o yaml | kubectl apply -f -

cp $GOPATH/src/github.com/openfaas-incubator/ofc-bootstrap/tmp/pub-cert.pem base/

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
                "name": "system-dashboard"
              }
            }
          },
          {
            "name": "payment_pointer",
            "valueFrom": {
              "configMapKeyRef": {
                "key": "payment_pointer",
                "name": "system-dashboard"
              }
            }
          }],
          "volumeMounts": [{
            "name": "pubcert",
            "mountPath": "/home/app/function/dist/pub-cert.pem",
            "subPath": "pub-cert.pem"
          }]
        }],
        "volumes": [{
          "name": "pubcert",
          "configMap": {
            "name": "system-dashboard",
            "items": [{
              "key": "pub-cert.pem",
              "path": "pub-cert.pem"
            }]
          }
        }]
      }
    }
  }
}'
