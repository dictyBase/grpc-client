# Kubernetes Deployment Guide: goldenbraid-list

This guide describes how to containerize and run the `goldenbraid-list` CLI tool in a Kubernetes cluster.

## 1. Containerization

Use the following `Dockerfile` to build a minimal image.

```dockerfile
# Build Stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /bin/goldenbraid-list ./cmd/goldenbraid-list

# Runtime Stage
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/goldenbraid-list /bin/goldenbraid-list
ENTRYPOINT ["/bin/goldenbraid-list"]
```

Build and push the image:
```bash
docker build -t your-registry/goldenbraid-list:latest .
docker push your-registry/goldenbraid-list:latest
```

## 2. Cross-Platform Building (macOS)

If you are building on a Mac (especially Apple Silicon M1/M2/M3), the default build will target `arm64`. Most Kubernetes clusters use `amd64` nodes. Use Docker Buildx to cross-compile:

```bash
# Build for amd64 Linux nodes
docker buildx build --platform linux/amd64 -t your-registry/goldenbraid-list:latest --push .
```

To optimize the `Dockerfile` for cross-compilation, you can use Go's built-in support for target architectures:

```dockerfile
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG TARGETOS TARGETARCH
WORKDIR /app
COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/goldenbraid-list ./cmd/goldenbraid-list
```

## 3. Run Options

### Ad-hoc Execution (kubectl run)
Run the tool immediately for one-off tasks.

```bash
kubectl run goldenbraid-list --rm -i -t 
  --image=your-registry/goldenbraid-list:latest 
  --restart=Never 
  -- --host=stock-api-service --filter="summary=~GoldenBraid"
```

### Kubernetes Job (One-time Task)
Use a Job for finite executions that require tracking.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: goldenbraid-list-job
spec:
  template:
    spec:
      containers:
      - name: goldenbraid-list
        image: your-registry/goldenbraid-list:latest
        args: ["--host", "stock-api-service", "--filter", "summary=~GoldenBraid"]
      restartPolicy: OnFailure
```

### CronJob (Scheduled Task)
Schedule the tool to run periodically (e.g., daily).

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: goldenbraid-list-daily
spec:
  schedule: "0 0 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: lister
            image: your-registry/goldenbraid-list:latest
            args: ["--host", "stock-api-service"]
          restartPolicy: OnFailure
```

### InitContainer (Pre-loading Data)
Fetch data before a main application starts.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-with-data
spec:
  template:
    spec:
      volumes:
      - name: data-volume
        emptyDir: {}
      initContainers:
      - name: fetcher
        image: your-registry/goldenbraid-list:latest
        command: ["/bin/sh", "-c"]
        args: ["/bin/goldenbraid-list --host=stock-api > /data/plasmids.txt"]
        volumeMounts:
        - name: data-volume
          mountPath: /data
      containers:
      - name: main-app
        image: nginx
        volumeMounts:
        - name: data-volume
          mountPath: /usr/share/nginx/html/data
```

## 3. Configuration

Since environment variables are not required, pass configuration directly via command-line flags:
- `--host`: gRPC server address
- `--port`: gRPC server port
- `--filter`: Stock API filter string
