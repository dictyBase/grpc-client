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
                - search
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
                - search
                - lookup
                - --name
                - "{{name}}"
                - --limit
                - "{{limit}}"
    EOF

# List all plasmids without filter in Kubernetes
run-search-list-all tag k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-list-all-
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
                - search
                - list-all
    EOF

# Fetch a single plasmid by identifier in Kubernetes
run-search-fetch tag identifier k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-fetch-
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
                - search
                - fetch
                - --identifier
                - "{{identifier}}"
    EOF

# Find annotations matching a filter in Kubernetes
run-annotation-find tag filter="" limit="20" cursor="0" k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-annotation-find-
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
                - annotation
                - find
                - --filter
                - "{{filter}}"
                - --limit
                - "{{limit}}"
                - --cursor
                - "{{cursor}}"
    EOF

# Find annotations by tag and ontology in Kubernetes
run-annotation-findbytag tag ontology="" tag_name="" limit="20" cursor="0" k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-annotation-findbytag-
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
                - annotation
                - findbytag
                - --ontology
                - "{{ontology}}"
                - --tag
                - "{{tag_name}}"
                - --limit
                - "{{limit}}"
                - --cursor
                - "{{cursor}}"
    EOF

# Retrieve annotation groups by identifier in Kubernetes
run-annotation-groupfind tag identifier ontology="" tag_name="" limit="20" cursor="0" k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-annotation-groupfind-
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
                - annotation
                - groupfind
                - --identifier
                - "{{identifier}}"
                - --ontology
                - "{{ontology}}"
                - --tag
                - "{{tag_name}}"
                - --limit
                - "{{limit}}"
                - --cursor
                - "{{cursor}}"
    EOF

# Delete an annotation by tag, identifier, and ontology in Kubernetes
run-annotation-remove tag tag_name identifier ontology k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-annotation-remove-
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
                - annotation
                - remove
                - --tag
                - "{{tag_name}}"
                - --identifier
                - "{{identifier}}"
                - --ontology
                - "{{ontology}}"
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

# Filter strains by type in Kubernetes
run-strain-filter tag strain_type="all" limit="20" cursor="0" k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-strain-filter-
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
                - strain
                - filter
                - --strain-type
                - "{{strain_type}}"
                - --limit
                - "{{limit}}"
                - --cursor
                - "{{cursor}}"
    EOF

# Fetch a single strain by identifier in Kubernetes
run-strain-fetch tag identifier k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-strain-fetch-
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
                - strain
                - fetch
                - --identifier
                - "{{identifier}}"
    EOF

# Create a feature annotation in Kubernetes
run-annofeat-create tag id name created_by="" synonyms="" properties="" k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    args=(
        - annofeat
        - create
        - --id
        - "{{id}}"
        - --name
        - "{{name}}"
    )
    if [ -n "{{created_by}}" ]; then
        args+=(--created-by "{{created_by}}")
    fi
    if [ -n "{{synonyms}}" ]; then
        args+=(--synonyms "{{synonyms}}")
    fi
    if [ -n "{{properties}}" ]; then
        args+=(--properties "{{properties}}")
    fi
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-annofeat-create-
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
    $(printf '                %s\n' "${args[@]}")
    EOF

# Retrieve a feature annotation by ID in Kubernetes
run-annofeat-get tag id k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    kubectl create -f - --kubeconfig "$kubeconfig" -o jsonpath='{.metadata.name}' <<EOF
    apiVersion: batch/v1
    kind: Job
    metadata:
      generateName: grpc-client-annofeat-get-
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
                - annofeat
                - get
                - --id
                - "{{id}}"
    EOF

# Get failure details for a job
job-debug name k8s_config="" k8s_namespace="dev" env="dev":
    #!/usr/bin/env bash
    set -euo pipefail
    kubeconfig=$(just resolve-kubeconfig "{{env}}" "{{k8s_config}}")
    echo "--- Pod Logs ---"
    kubectl logs job/{{name}} --kubeconfig "$kubeconfig" -n {{k8s_namespace}} || true
    echo "--- Job Description ---"
    kubectl describe job/{{name}} --kubeconfig "$kubeconfig" -n {{k8s_namespace}}
