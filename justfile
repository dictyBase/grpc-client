# Justfile for goldenbraid-list

name := "goldenbraid-list"
namespace := "dictybase"
github_user := "sba964"
platform := "linux/amd64"
platform_multi := "linux/amd64,linux/arm64"

image := namespace + "/" + name
ghcr_image := "ghcr.io/" + image

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

# List GoldenBraid plasmids in dev cluster
run-list tag filter="summary=~GoldenBraid":
    #!/usr/bin/env bash
    set -euo pipefail
    export KUBECONFIG=$(k3d kubeconfig write k3d-dev-cluster)
    kubectl create -f - <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: goldenbraid-list-
      namespace: dev
    spec:
      ttlSecondsAfterFinished: 200
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: goldenbraid-list
              image: {{ghcr_image}}:{{tag}}
              args:
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
      generateName: goldenbraid-lookup-
      namespace: dev
    spec:
      backoffLimit: 0
      ttlSecondsAfterFinished: 200
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: goldenbraid-lookup
              image: {{ghcr_image}}:{{tag}}
              args:
                - lookup
                - --name
                - "{{name}}"
                - --limit
                - "{{limit}}"
    EOF

# Wait for a Kubernetes job to complete, fail, or detect stuck pods.
# Delegates to the goldenbraid-list wait-job subcommand (fp-go implementation).
wait-job name namespace="dev" timeout="60s":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(k3d kubeconfig write k3d-dev-cluster)
    go run ./cmd/goldenbraid-list/ wait-job --name {{name}} --namespace {{namespace}} --timeout {{timeout}} --kubeconfig "$kubeconfig"

# Get the logs for a specific job
job-logs name namespace="dev":
    #!/usr/bin/env bash
    export KUBECONFIG=$(k3d kubeconfig write k3d-dev-cluster)
    kubectl logs job/{{name}} -n {{namespace}}

# Get failure details for a job
job-debug name namespace="dev":
    #!/usr/bin/env bash
    export KUBECONFIG=$(k3d kubeconfig write k3d-dev-cluster)
    echo "--- Pod Logs ---"
    kubectl logs job/{{name}} -n {{namespace}} || true
    echo "--- Job Description ---"
    kubectl describe job/{{name}} -n {{namespace}}
