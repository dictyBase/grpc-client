# Justfile for grpc-client

name := "grpc-client"
namespace := "dictybase"
github_user := "sba964"
platform := "linux/amd64"
platform_multi := "linux/amd64,linux/arm64"

image := namespace + "/" + name
ghcr_image := "ghcr.io/" + image

[private]
resolve-kubeconfig env="dev" k8s_config="":
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -n "{{k8s_config}}" ]; then
        printf '%s\n' "{{k8s_config}}"
    elif [ "{{env}}" = "dev" ]; then
        kubeconfig=$(mktemp)
        k3d kubeconfig get k3d-dev-cluster > "$kubeconfig"
        printf '%s\n' "$kubeconfig"
    else
        echo "ERROR: k8s_config must be provided for {{env}} environment" >&2
        exit 1
    fi

# Build the docker image for the target platform
build tag="latest":
    docker buildx build --platform {{platform}} -t {{image}}:{{tag}} .

# Build and push the docker image
push tag="latest":
    docker buildx build --platform {{platform}} -t {{image}}:{{tag}} --push .

# Build for GitHub Container Registry
build-ghcr tag="latest":
    docker buildx build --platform {{platform}} -t {{ghcr_image}}:{{tag}} .

# Push to GitHub Container Registry
push-ghcr tag="latest":
    echo $GITHUB_REGISTRY_TOKEN | docker login ghcr.io -u {{github_user}} --password-stdin
    docker buildx build --platform {{platform}} -t {{ghcr_image}}:{{tag}} --push .

# Push multi-arch image (amd64 + arm64) to GitHub Container Registry
push-ghcr-multi tag="latest":
    echo $GITHUB_REGISTRY_TOKEN | docker login ghcr.io -u {{github_user}} --password-stdin
    docker buildx build --platform {{platform_multi}} -t {{ghcr_image}}:{{tag}} --push .

# List images
list:
    docker images | grep {{image}}

# List GoldenBraid plasmids in Kubernetes
run-list tag filter="summary=~GoldenBraid" k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-list-
      namespace: {{k8s_namespace}}
    spec:
      ttlSecondsAfterFinished: 200
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: grpc-client
              image: {{ghcr_image}}:{{tag}}
              args:
                - plasmid
                - list
                - --filter
                - "{{filter}}"
    EOF

# Look up a GoldenBraid plasmid by exact name in dev cluster
run-lookup tag name limit="3":
    #!/usr/bin/env bash
    set -euo pipefail
    export KUBECONFIG=$(k3d kubeconfig write k3d-dev-cluster)
    kubectl create -f - <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-lookup-
      namespace: dev
    spec:
      backoffLimit: 0
      ttlSecondsAfterFinished: 200
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: grpc-client-lookup
              image: {{ghcr_image}}:{{tag}}
              args:
                - plasmid
                - lookup
                - --name
                - "{{name}}"
                - --limit
                - "{{limit}}"
    EOF

# Wait for a Kubernetes job to complete, fail, or detect stuck pods.
# Delegates to the grpc-client wait-job subcommand (fp-go implementation).
wait-job name k8s_config="" k8s_namespace="dev" env="dev" timeout="60s":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    go run ./cmd/grpc-client/ wait-job --name {{name}} --namespace {{k8s_namespace}} --timeout {{timeout}} --kubeconfig "$kubeconfig"

# Get the logs for a specific job
job-logs name k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl logs job/{{name}} --kubeconfig "$kubeconfig" -n {{k8s_namespace}}

# Get failure details for a job
job-debug name k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    echo "--- Pod Logs ---"
    kubectl logs job/{{name}} --kubeconfig "$kubeconfig" -n {{k8s_namespace}} || true
    echo "--- Job Description ---"
    kubectl describe job/{{name}} --kubeconfig "$kubeconfig" -n {{k8s_namespace}}
