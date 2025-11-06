default:
  @just --list

up: start-kind-cluster \
build-deps \
install-boxer \
install-ingress-controller \
instll-keycloak \
wait-for-services \
create-ingress \
configure-keycloak \
bootstrap

fresh: stop up

stop:
    kind delete cluster --name kind

start-kind-cluster:
    kind create cluster --name kind --config=integration_tests/kind.yaml

check-kind-cluster:
    kubectl cluster-info
    kubectl get nodes
    kind get kubeconfig --name kind

build-deps:
    helm dependency build ./integration_tests/helm/setup

key := `openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | fold -w 16 | head -n 1`
install-boxer:
    helm upgrade --install --namespace default integration-tests integration_tests/helm/setup \
      --set boxer-issuer.issuer.replicas=1 \
      --set-literal 'boxer-validator-nginx.validator.config.tokenSettings.keys={"default": "{{key}}"}' \
      --set 'boxer-issuer.issuer.config.listenIp=0.0.0.0' \
      --set 'boxer-issuer.issuer.config.logLevel=debug' \
      --set 'boxer-issuer.issuer.config.backend.kubernetes.resourceOwnerLabel=application/boxer-issuer' \
      --set 'boxer-validator-nginx.validator.config.listenIp=0.0.0.0' \
      --set 'boxer-validator-nginx.validator.config.backend.kubernetes.resourceOwnerLabel=application/boxer-validator-nginx' \
      --set boxer-validator-nginx.validator.replicas=1

wait-for-services:
    kubectl rollout status deployment/boxer-issuer --timeout=180s
    kubectl rollout status deployment/boxer-validator-nginx --timeout=180s
    kubectl rollout status deployment/ingress-nginx-controller --namespace ingress-nginx --timeout=180s
    kubectl rollout status statefulset/keycloak-keycloakx --timeout=180s

install-ingress-controller:
    kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml

create-ingress:
    # Wait a bit for ingress controller to be ready to accept rules
    sleep 10
    # Create ingress rules for boxer-issuer and boxer-validator-nginx
    kubectl apply -f ./integration_tests/ingress.yaml

bootstrap:
    kubectl apply -f ./integration_tests/bootstrap/bootstrap.yaml

## Check: run this command in the boxer-issuer container to verify
## curl -H "Authorization: Bearer your_token" http://localhost/api/v1/identity_provider/oidc/rooatr/oidc/root

instll-keycloak:
    helm upgrade --install keycloak oci://ghcr.io/codecentric/helm-charts/keycloakx \
      --set keycloak.username=admin \
      --set keycloak.password=admin \
      --values ./integration_tests/keycloak.yaml

configure-keycloak:
    # Wait a bit for Keycloak to be ready to accept admin commands
    sleep 10

    # Create realm, client, and user for tests
    docker run --rm --network=host -v $(pwd)/integration_tests/terraform/keycloak:/tofu --workdir /tofu \
      ghcr.io/opentofu/opentofu:latest init
    docker run --rm --network=host -v $(pwd)/integration_tests/terraform/keycloak:/tofu --workdir /tofu \
      ghcr.io/opentofu/opentofu:latest plan
    docker run --rm --network=host -v $(pwd)/integration_tests/terraform/keycloak:/tofu --workdir /tofu \
      ghcr.io/opentofu/opentofu:latest apply -auto-approve
