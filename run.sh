#!/bin/bash +x
set -euo pipefail

# Parse command line arguments
CICD=false
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --cicd) CICD=true ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

# Install required tools on Mac
if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install docker k6 helm minikube kubectl kubectx go argocd
elif [[ "$CICD" == true ]]; then
    curl -LO https://github.com/kubernetes/minikube/releases/latest/download/minikube-linux-amd64
    sudo install minikube-linux-amd64 /usr/local/bin/minikube && rm minikube-linux-amd64
    curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
    sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
    rm argocd-linux-amd64
    sudo apt-get update; sudo apt-get install -y expect
fi

function wait_for_healthy_apps() {
    local timeout=$1
    local start_time=$(date +%s)

    while true; do
        current_time=$(date +%s)
        elapsed=$((current_time - start_time))
        
        if [ $elapsed -ge $timeout ]; then
            echo "Timeout reached after $((timeout/60)) minutes - exiting"
            return 1
        fi
        
        echo "Checking application health status... ${elapsed}s elapsed"
        still_waiting=false
        while IFS= read -r line; do
            name=$(echo "$line" | awk '{print $1}')
            status=$(echo "$line" | awk '{print $5}')
            health=$(echo "$line" | awk '{print $6}')
            if [ "$health" != "Healthy" ] || [ "$status" == "Unknown" ]; then
                echo "App $name is in health:$health status:$status state"
                still_waiting=true
            fi
        done < <(argocd app list | tail -n +2)
        argocd app list | tail -n +2
        if [ "$still_waiting" == false ]; then
            echo "All applications are healthy!"
            return 0
        fi
        sleep 5
    done
}

# Setup Kubernetes environment/ comment this out if not using minikube
minikube start
minikube addons enable registry
minikube addons enable ingress
minikube addons enable metrics-server

# # Verify kubectl
# sleep 30
# kubectl get pods

# use tofu instead

# # Install ArgoCD
if ! kubectl get namespace argocd &>/dev/null; then
    echo "Creating ArgoCD namespace and installing ArgoCD..."
    kubectl create namespace argocd
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
    echo "sleeping for 40 seconds to allowing argo to deploy and start  ... .. ."
    sleep 40

else
    echo "ArgoCD namespace already exists, continuing..."
fi

# bootstrap the cluster with argo applications
kubectl apply -f charts/bootstrap/

# # Build and load app image
docker build . -t sample-app
minikube image load sample-app

# echo "sleeping for 30 seconds to allowing argo to deploy and start ... .. ."
# sleep 30

# Get ArgoCD admin password and setup port-forward
echo "Waiting for ArgoCD secret to be created..."
timeout=120
start_time=$(date +%s)

while true; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))
    
    if [ $elapsed -ge $timeout ]; then
        echo "Timeout waiting for ArgoCD secret"
        exit 1
    fi
    
    if kubectl -n argocd get secret argocd-initial-admin-secret &>/dev/null; then
        echo "Secret found!"
        break
    fi
    
    echo "Secret not found, waiting... ${elapsed}s elapsed"
    sleep 5
done

ARGO_SECRET=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)

echo "Waiting for argocd-server to be ready..."
sleep 20
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=120s
# if [[ "$CICD" == true ]]; then
#     echo "CICD is true, skipping argocd login"
#     # this is busted
#     argocd login $(minikube service argocd-server --url -n argocd) --username admin --password ${ARGO_SECRET} --insecure
# else
#     kubectl port-forward svc/argocd-server -n argocd 8080:443 &
#     argocd login localhost:8080 --username admin --password ${ARGO_SECRET} --insecure

# fi
sleep 20
kubectl port-forward svc/argocd-server -n argocd 8080:443 &
sleep 5
argocd login localhost:8080 --username admin --password ${ARGO_SECRET} --insecure


timeout=600  # 10 minutes
wait_for_healthy_apps $timeout

kubectl apply -f charts/monitoring/
sleep 10

# deploy the app through argocd
kubectl apply -f charts/applications.yaml
wait_for_healthy_apps $timeout
if [[ "$CICD" == true ]]; then
    echo "shutting down the argocd port-forward"
    pgrep -f "kubectl port-forward svc/argocd-server -n argocd 8080:443" |  xargs kill
    echo ""
fi

# Cleanup
unset ARGO_SECRET

## 
cat <<EOF

You can now port forward or turn on minikube tunnel to access the application through ingress.


To port forward:
kubectl port-forward svc/backend  -n sample-app 8080:443
curl -X GET http://localhost:8080/

or with ingress
minikube tunnel
curl -i -L  https://example.com/hello -k

Check out the readme for more detailed information - 

enjoy!

EOF

