# Sample app example

> [!CAUTION]
> **Disclaimer: This repo is a sample app displaying a wide range of examples. It's a playground and always a work in progress. Please do not take any of this code as production ready.**

Many pieces need to come together to create successful distributed systems. *Testing, builds, deployments, observability, and security* are all critical in order to build a robust system. This repo is a playground to help build a sample application that can be a reference or used as a starting point for your own projects.

Here is a quick checklist of the components that are covered in this repo.

```
[X] Application
  [X] - Application metrics
  [X] - Testing (details in the testing section)
  [X] - Ingress + HTTPS
  [-] - Review/Improve error handling
[X] - Testing
  [X] - Unit tests
  [X] - Integration tests
  [X] - Load tests
  [-] - Improve Security tests
[X] - Building
  [X] - CI pipelines
  [X] - Docker builds
  [X] - Local registry
[X] - Observability
  [X] - Grafana
  [X] - Prometheus
  [X] - OpenTelemetry examples
  [X] - Logging
  [X] - Metrics
  [-] - Tracing
  [-] - Alerting
  [-] - Dashboards
[X] - Deployable
  [X] - Infrastructure Bootstrap
  [X] - Helm chart
  [X] - Argocd
  [X] - Kubernetes yaml
  [X] - Kubernetes Operator examples
  [X] - Docker
  [-] - Improve secret management
```


# Whats in the go app in this repo?

The app itself is simple. It prints `Hello, World!` on the `/hello` endpoint, however there is more code examples in here relating to observability and other goodies. 

# Getting started

> [!TIP]
> If you do not need a step-by-step, feel free to just run the `./run.sh` and skip to the port-forwarding and testing section.
> If you want a step-by-step continue in the section below.
>
> Usage:
>
>      
>      ./run.sh        (--cicd [optional flag used in CI/CD])
>      

## Prerequisites

- You will need these tools
  - [Docker](https://docs.docker.com/get-docker/)
  - [Go](https://go.dev/doc/install)
  - [Kubernetes](https://kubernetes.io/docs/tasks/tools/)
  - [Argocd](https://argoproj.github.io/argo-helm/)
  - [k6](https://k6.io/open-source/)

For y'all mac users, you can use Homebrew to install the needed tools

```bash
brew install docker k6 helm minikube kubectl kubectx go argocd
```

Steps to get your Kubernetes environment setup => 
1. Start minikube:

```bash
minikube start; minikube addons enable registry; minikube addons enable ingress; minikube addons enable metrics-server
```

2. Confirm kubectl is working:

```bash
kubectl get pods
```

3. Install infrastructure apps by using **one of the two methods below.**
- Automated method - *Argocd (preferred method)*
  ```bash
  kubectl create namespace argocd
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
  kubectl apply -f charts/bootstrap/
  # or
  # helm repo add argo https://argoproj.github.io/argo-helm ; helm repo update # or by using helm
  # helm upgrade -i my-argo argo/argo-cd --namespace=argocd --create-namespace --wait
  ```

> [!NOTE]
> OPTIONAL - *Install Prometheus/grafana/argocd manually through helm instead of through argocd (less cool but okay)*
> ```bash
> helm repo add bitnami https://charts.bitnami.com/bitnami
> helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
> helm repo add grafana https://grafana.github.io/helm-charts
> helm repo add argo https://argoproj.github.io/argo-helm
> helm repo update
> # helm upgrade -i my-argo argo/argo-cd --namespace=argocd --create-namespace --wait #if needed argo can be added too 
> helm install grafana grafana/grafana
> ```

In the future of this project, only argocd should be needed.

Great! The cluster is now up and ready to deploy apps through argpcd.

# Build 

Suggestions for builds.
- Make containers as small as you can. This will allow for quicker builds and scaling. Saving on cost and reduce the time to deploy.
- Use multi-stage builds to reduce the size of the final image and remove unnecessary files.

Build the image and confirm the local minikube-docker registry is running.
```bash
docker build . -t sample-app
# you can test it locally with Docker run else skip
#docker run --rm -p 8080:8080 sample-app
# confirm you can see the image in the minikube Docker registry (for local use only)
# you may need to load it in
minikube image load sample-app
minikube image ls | grep sample-app
```

# Deploy - Kubernetes/Argocd

The deployment is split into two sections. 
1. Bootstrap the cluster (done)
2. Deploy the application itself

## Bootstrap Deploy

Your cluster should already be bootstrapped with argocd and any other clusterwide tools needed.

```bash
argocd app list -o name | xargs -n 1 argocd app get --refresh
```

Review in Argo UI, or you can use the CLI
```bash
# Get the admin password for local argo - you may need to trigger the sync at this time till i fix it
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Access via port-forward (default username: admin)
kubectl port-forward svc/argocd-server -n argocd 8080:443

```

## Application Deploy

This also takes place in the argocd config. See `./charts/bootstrap/applications.yaml` for the deployment of the source in this repo. 

> [!NOTE]
> If you didnt want to use argocd and needed to deploy the app manually you can use the below code:

> ```bash
> helm repo add bitnami https://charts.bitnami.com/bitnami
> helm repo update
> helm dependency update charts/sample-app
> helm upgrade -i sample-app charts/sample-app --namespace=sample-app --create-namespace --wait
> ```

# Testing

unit tests:
```bash
go test -v ./...
```

manual testing endpoints:
```bash
kubectl port-forward svc/backend 8080:8080 -n sample-app
# or use nginx ingress

# Make some requests to generate metrics
curl http://localhost:8080/hello
curl -H "Fail: true" http://localhost:8080/hello
curl -X POST http://localhost:8080/hello
curl -H "Delay: 5s" localhost:8080/hello

# Then check the metrics
curl http://localhost:8080/metrics
```

```bash
while true; do sleep .5 ; curl localhost:8080/hello; done
```

integration tests:
- using k6

```bash
k6 run --vus 10 ./test/k6.js
k6 run -e BASE_URL=http://localhost:8080 test/k6.js

```

# Ingress

nginx nginx ingress controller is used to expose the application using https. This is done by creating a service of type `LoadBalancer` and then using the ingress controller to route traffic to the service.

do the load tests over the ingress address

```bash
# you may need to add example.com to your /etc/hosts file for this to work
# sudo echo "127.0.0.1 example.com" >> /etc/hosts
curl -i -L  https://example.com/hello -k
```

```bash
k6 run --vus 10 ./test/k6.js
k6 run -e BASE_URL=http://localhost:8080 test/k6.js

```

# Observability

For this example I use Prometheus and Grafana to display open telemetry metrics from the go application. They are deployed using the Kubernetes operator and CRDs.

## Metrics

Metrics are displayed at the `/metrics` endpoint in our GO application using the Prometheus open telemetry library.
The following metrics are currently available or areas for future development:

* `http_requests_total`: The total number of HTTP requests.
* `http_request_duration_seconds`: The duration of HTTP requests in seconds.
* `http_request_size_bytes`: The size of HTTP requests in bytes [future feature].
* `http_response_size_bytes`: The size of HTTP responses in bytes [future feature].

## Alerting - 4 golden signals 

> [!TIP]
> Try and stick to the 4 golden signals as much as you can. If you application has some specific case, try and related it back to one of the signals. See where it could fit. This can help reduce alert fatigue. 

In these examples, Prometheus is configured to send alerts based off the following alerting rules:
* `HighErrorRate`: Alerts when the error rate is over 1% for 1 minute.
* `HighLatency`: Alerts when the latency is over 1 second for 1 minute.
* `HighRequestVolume`: Alerts when the request volume is over 100 requests per second for 1 minute.

Alerts are deployed on a per app basis in the helm chart, similar to the dashboard example.

## Dashboarding - Visualizing Metrics

The concept to good dashboard is to provide a default dashboard with the common values to review. For example,

- CPU/Memory usage
- Application Metrics
- Kubernetes Metrics
- Scaling Details
- etc..

[Dashboard are deployed](./charts/sample-app/templates/dashboard.yaml) to the [Grafana operator](https://grafana.com/docs/grafana-cloud/developer-resources/infrastructure-as-code/grafana-operator/manage-dashboards-argocd/) on a per-service basis. This is a default dashboard which can be edited and customized to your needs.


## Observability Examples

Port-forward the Prometheus and Grafana services:

```bash
kubectl port-forward svc/prometheus-operated -n monitoring 9090:9090
kubectl port-forward svc/grafana-service 3000:3000 -n monitoring
```

Head over to the `Query` button. Type in one of the below examples to get started. Click the graph button to see pretty pictures.

Most of the query examples are related to the [4 golden signals](https://sre.google/sre-book/monitoring-distributed-systems/):

Errors and tracing:

```
http_requests_total{namespace="sample-app"}

# Request rate (per second) over the last 5 minutes for a namespace :
rate(http_requests_total{namespace="sample-app"}[5m])

# Total requests per endpoint, showing requests/second:
sum by (endpoint) (rate(http_requests_total[5m]))

# Error rate percentage:
sum(rate(http_requests_total{code=~"5.."}[5m])) 
  / 
sum(rate(http_requests_total[5m])) 
* 100

# Top 5 busiest endpoints:
topk(5, sum by (endpoint) (rate(http_requests_total[5m])))


# Request rate grouped by status code:
sum by (code) (rate(http_requests_total[5m]))

# Success vs failure comparison:
sum by (status_class) (
  rate(http_requests_total{code=~"2.."}[5m]) or 
  rate(http_requests_total{code=~"[45].."}[5m])
)

# Moving average over 1 hour to smooth out spikes:
avg_over_time(rate(http_requests_total[5m])[1h:5m])
```

Latency:
```promql
# Request latency distribution as a heatmap:
rate(http_request_duration_seconds_bucket{endpoint="/hello"}[5m])

# average request duration:
rate(http_request_duration_seconds_sum{endpoint="/hello"}[5m]) 
  / 
rate(http_request_duration_seconds_count{endpoint="/hello"}[5m])

#multiple percentiles at (50th, 90th, 95th, 99th):
## this isnt fully working yet
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{endpoint="/hello"}[5m]))
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{endpoint="/hello"}[5m]))
histogram_quantile(0.90, rate(http_request_duration_seconds_bucket{endpoint="/hello"}[5m]))
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{endpoint="/hello"}[5m]))
```

Kubernetes and operational type examples - still working some of these out in minikube
```
sum(container_cpu_usage_seconds_total{container!="backend"}) by (pod)

sum(container_memory_usage_bytes{container!="backend"}) by (pod)

#show the number of pods per deployment
sum(kube_pod_status_phase) by (phase)

#container restarts?
sum(kube_pod_container_status_restarts_total) by (pod)
```


# Cleanup

To delete and start over, delete the minikube instance.

```bash
minikube stop; minikube delete
```

