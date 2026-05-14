
# grpc-client

CLI and workflow executor for dictyBase gRPC services, deployed as Kubernetes Jobs orchestrated by Dagu DAGs.

## Contents

- [Prerequisites](#prerequisites)
- [Install](#install)
  - [Build binary](#build-binary)
  - [Build Docker image](#build-docker-image)
  - [Publish to GHCR](#publish-to-ghcr)
- [Commands](#commands)
  - [search list-all](#search-list-all)
  - [search list](#search-list)
  - [search lookup](#search-lookup)
  - [search fetch](#search-fetch)
  - [strain fetch](#strain-fetch)
  - [strain filter](#strain-filter)
  - [annotation find](#annotation-find)
  - [annotation findbytag](#annotation-findbytag)
  - [annotation groupfind](#annotation-groupfind)
  - [annotation remove](#annotation-remove)
  - [annofeat create](#annofeat-create)
  - [annofeat get](#annofeat-get)
  - [wait-job](#wait-job)
- [Kubernetes](#kubernetes)
  - [Dagu Workflows](#dagu-workflows)
  - [just targets](#just-targets)

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Dagu](https://github.com/dagu-dev/dagu) — DAG workflow runner
- [just](https://github.com/casey/just) — command runner
- [kubectl](https://kubernetes.io/docs/tasks/tools/) — Kubernetes CLI
- [k3d](https://k3d.io/) — local dev cluster (for `ENV=dev`)
- [Docker](https://docs.docker.com/) with [Buildx](https://docs.docker.com/buildx/working-with-buildx/) — image builds and multi-arch support

## Install

### Build binary

```bash
go build -o grpc-client ./cmd/grpc-client
```

### Build Docker image

Builds a local image for `linux/amd64` without pushing.

```bash
just build          # tags as dictybase/grpc-client:latest
just build tag=1.2  # tags as dictybase/grpc-client:1.2
```

To build a GHCR-tagged image locally without pushing:

```bash
just build-ghcr          # tags as ghcr.io/dictybase/grpc-client:latest
just build-ghcr tag=1.2
```

### Publish to GHCR

All images are published to **GitHub Container Registry (GHCR)** at:

```
ghcr.io/dictybase/grpc-client:<tag>
```

Kubernetes Jobs pull from this registry. Before pushing you must authenticate.

#### Before you begin

1. Create a GitHub Personal Access Token (PAT) with the `write:packages` scope at
   <https://github.com/settings/tokens>.
2. Export it as an environment variable:

   ```bash
   export GITHUB_REGISTRY_TOKEN=<your-pat>
   ```

#### Push commands

```bash
just push-ghcr            # single-arch (linux/amd64), latest tag
just push-ghcr tag=1.2    # single-arch, custom tag
just push-ghcr-multi      # multi-arch (linux/amd64 + linux/arm64), latest tag
just push-ghcr-multi tag=1.2
```

Both targets authenticate with `docker login ghcr.io` using `GITHUB_REGISTRY_TOKEN` automatically.

> **Multi-arch builds** require a Docker Buildx builder with multi-platform drivers enabled. Create one if you have not already:
> ```bash
> docker buildx create --use --name multiarch --driver docker-container
> ```

## Commands

gRPC host/port are auto-detected from environment variables or passed via `--host`/`--port` flags.

| Service | Host env | Port env |
|---------|----------|----------|
| Stock API | `STOCK_API_SERVICE_HOST` | `STOCK_API_SERVICE_PORT` |
| Annotation API | `ANNOTATION_API_SERVICE_HOST` | `ANNOTATION_API_SERVICE_PORT` |
| Feature Annotation API | `ANNO_FEAT_API_SERVICE_HOST` | `ANNO_FEAT_API_SERVICE_PORT` |

### search list-all

List all plasmids without filter, paginated 30 at a time. Prints the total count and first 10 records.

```
grpc-client search list-all [--host HOST] [--port PORT]
```

### search list

List plasmids matching a filter. Default filter: `summary=~GoldenBraid`.

```
grpc-client search list [--host HOST] [--port PORT] [--filter FILTER]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--filter` | `summary=~GoldenBraid` | Stock API filter string; env: `STOCK_API_FILTER` |

Filter syntax: [modware-stock filtering](https://github.com/dictyBase/modware-stock#filtering) — supports `;` (AND), `,` (OR), string/date/array operators.

### search lookup

Look up a plasmid by exact name using a `plasmid_name===` filter.

```
grpc-client search lookup --name NAME [--host HOST] [--port PORT] [--limit N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | *(required)* | Exact plasmid name (e.g. `pDGB3alpha1`) |
| `--limit` | `3` | Max results |

Filter syntax: [modware-stock filtering](https://github.com/dictyBase/modware-stock#filtering).

### search fetch

Fetch a single plasmid by its identifier.

```
grpc-client search fetch --identifier ID [--host HOST] [--port PORT]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--identifier` / `-i` | *(required)* | Plasmid identifier (e.g. `DBP0000001`) |

### strain fetch

Fetch a single strain by its identifier.

```
grpc-client strain fetch --identifier ID [--host HOST] [--port PORT]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--identifier` / `-i` | *(required)* | Strain identifier (e.g. `DBS0000001`) |

### strain filter

List strains by type using `ontology==dicty_strain_property;tag==` filters. Allowed types: `REMI-seq`, `general strain`, `bacterial strain`, `all`.

```
grpc-client strain filter [--strain-type TYPE] [--limit N] [--cursor N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--strain-type` / `-st` | `all` | Strain type filter |
| `--limit` / `-l` | `10` | Results per page |
| `--cursor` | `0` | Pagination offset |

Filter syntax: [modware-stock filtering](https://github.com/dictyBase/modware-stock#filtering).

### annotation find

Find annotations matching a filter string.

```
grpc-client annotation find [--filter FILTER] [--limit N] [--cursor N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--filter` | `""` | Filter string (e.g. `entry_id===DDB_G123;ontology===cellular_component`) |
| `--limit` / `-l` | `10` | Results per page |
| `--cursor` | `0` | Pagination offset |

Filter syntax: [modware-annotation filtering](https://github.com/dictyBase/modware-annotation#filtering) — supports `;` (AND), `,` (OR), string/numeric/boolean operators. Filterable fields: `entry_id`, `value`, `created_by`, `tag`, `ontology`, `version`, `rank`, `is_obsolete`.

### annotation findbytag

Find annotations filtered by ontology and tag using `tag===;ontology===` filter.

```
grpc-client annotation findbytag --ontology ONTOLOGY --tag TAG [--limit N] [--cursor N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--ontology` | `""` | Ontology name (e.g. `cellular_component`, `biological_process`) |
| `--tag` | `""` | Tag/term in ontology (e.g. `GO:0005634`) |
| `--limit` / `-l` | `10` | Results per page |
| `--cursor` | `0` | Pagination offset |

Filter syntax: [modware-annotation filtering](https://github.com/dictyBase/modware-annotation#filtering).

### annotation groupfind

Retrieve annotation groups by identifier using `entry_id===` filter, optionally narrowed by tag and ontology.

```
grpc-client annotation groupfind --identifier ID [--ontology ONTOLOGY] [--tag TAG] [--limit N] [--cursor N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--identifier` / `-i` | *(required)* | Entry identifier (e.g. `DDB_G123`) |
| `--ontology` | `""` | Optional ontology filter |
| `--tag` | `""` | Optional tag filter |
| `--limit` / `-l` | `100` | Results per page |
| `--cursor` | `0` | Pagination offset |

Filter syntax: [modware-annotation group filtering](https://github.com/dictyBase/modware-annotation#group-filtering).

### annotation remove

Delete an annotation by tag, identifier, and ontology.

```
grpc-client annotation remove --tag TAG --identifier ID --ontology ONTOLOGY
```

| Flag | Default | Description |
|------|---------|-------------|
| `--tag` | *(required)* | Tag/term (e.g. `GO:0005634`) |
| `--identifier` / `-i` | *(required)* | Entry identifier |
| `--ontology` | *(required)* | Ontology name |

### annofeat create

Create a new feature annotation.

```
grpc-client annofeat create --id ID --name NAME [--created-by EMAIL] [--synonyms LIST] [--properties KV]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--id` | *(required)* | Feature ID (e.g. `DDB_G0285425`) |
| `--name` | *(required)* | Feature name |
| `--created-by` | env: `ANNO_FEAT_CREATED_BY` | Creator email |
| `--synonyms` | `""` | Comma-separated synonyms (e.g. `test1,test2`) |
| `--properties` | `""` | Comma-separated `key=value` pairs (e.g. `description=Test,note=Info`) |

### annofeat get

Retrieve a feature annotation by ID.

```
grpc-client annofeat get --id ID
```

| Flag | Default | Description |
|------|---------|-------------|
| `--id` | *(required)* | Feature ID to retrieve |

### wait-job

Poll a Kubernetes Job until complete, failed, or stuck. Polls every 5s.

```
grpc-client wait-job --name NAME [--namespace NS] [--timeout DURATION] [--kubeconfig PATH]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | *(required)* | Job name |
| `--namespace` | `dev` | Kubernetes namespace |
| `--timeout` | `60s` | Max wait duration |
| `--kubeconfig` | env: `KUBECONFIG` | Path to kubeconfig |

## Kubernetes

Both Dagu DAGs and `just` targets deploy to Kubernetes. Dagu workflows orchestrate multi-step submit → wait → log pipelines; `just` targets are the low-level Job runners that Dagu calls internally. All Jobs pull the image from `ghcr.io/dictybase/grpc-client:<tag>`.

### Dagu Workflows

Multi-step DAG workflows orchestrate submit → wait → log retrieval with failure handlers.

| DAG | Purpose |
|-----|---------|
| `dagu-lookup.yaml` | Lookup plasmid → wait → fetch logs |
| `dagu-goldenbraid-list.yaml` | List GoldenBraid plasmids |
| `dagu-annotation-find.yaml` | Find annotations by filter |
| `dagu-annotation-findbytag.yaml` | Find annotations by tag + ontology |

### Running a workflow

Dagu reads parameters from `params.values` and passes them as environment variables to each step. Use the `+` character in place of spaces within parameter values.

```bash
# Lookup a plasmid by name
dagu start dagu-lookup.yaml TAG=latest NAME=pDGB3alpha1

# List GoldenBraid plasmids
dagu start dagu-goldenbraid-list.yaml TAG=latest FILTER=summary=~GoldenBraid

# Find annotations by filter (use + for whitespace)
dagu start dagu-annotation-find.yaml \
  TAG=latest \
  FILTER='ontology===cellular_component;tag===GO:0005634'

# Find annotations by tag (use + for multi-word values)
dagu start dagu-annotation-findbytag.yaml \
  TAG=latest \
  ONTOLOGY=cellular_component \
  TAG_NAME=GO:0005634
```

### How `+` works

Dagu splits parameters on spaces, so `FILTER=summary=~GoldenBraid` with a literal space would be misparsed. The workflow steps use `tr '+' ' '` to convert `+` back to spaces:

```yaml
# From dagu-annotation-findbytag.yaml — converts '+' back to spaces
_tag_name=$(printenv TAG_NAME | tr '+' ' ')
```

This means `TAG_NAME=general+strain` becomes `general strain` at runtime.

### Failure handling

If any step fails, the `handler_on.failure` block runs `job-debug` to dump pod logs and the job description, or reports that submission failed before a name was captured.

### just targets

`just` targets create Kubernetes Jobs directly.

```bash
# Plasmids
just run-lookup latest pDGB3alpha1 3
just run-search-list-all latest
just run-search-fetch latest DBP0000001

# Strains
just run-strain-fetch latest DBS0000001
just run-strain-filter latest "general strain" 20 0

# Annotations
just run-annotation-find latest "ontology===cellular_component" 20 0
just run-annotation-findbytag latest cellular_component GO:0005634 20 0
just run-annotation-groupfind latest DDB_G123 cellular_component GO:0005634
just run-annotation-remove latest GO:0005634 DDB_G123 cellular_component

# Feature annotations
just run-annofeat-create latest DDB_G0285425 "Test Feature"
just run-annofeat-get latest DDB_G0285425
```

Each Job self-destructs after 200s (`ttlSecondsAfterFinished`).

Job debugging:

```bash
just wait-job <job-name>
just job-logs <job-name>
just job-debug <job-name>
```
