# L2SM-DNS

**l2sm-dns** is a DNS service component for the [l2sm](https://github.com/Networks-it-uc3m/L2S-M) project. It provides a standard layout for Go-based CLI applications and services, enabling you to manage DNS entries via a gRPC server that integrates with CoreDNS and deploys seamlessly on Kubernetes.

## Project Structure

The repository is organized as follows:

- **cmd/**  
  - **server/**: Contains the gRPC server implementation (`server.go` and `main.go`) that listens on port `8081` and integrates with CoreDNS.
- **api/**  
  - **v1/dns/**: Holds the API definitions (e.g., `dns.proto`) and generated gRPC code (`dns.pb.go`, `dns_grpc.pb.go`).
- **pkg/**  
  - **coredns-manager/**: Implements the CoreDNSManager used to update the CoreDNS ConfigMap dynamically.
- **config/**  
  - **rbac/**: Kubernetes RBAC manifests for service accounts, roles, and role bindings.
  - **dev/**: Patches for local development (e.g., deleting or patching CoreDNS deployments).
  - **server/**: Kubernetes manifests (deployment, service, configmap) to run the DNS service in a cluster.
  - **default/**: Default kustomize configurations.
- **test/**  
  - Contains end-to-end and integration tests, along with a sample DNS client (`client.go`) and configuration (`config.yaml`).
- **deployments/**  
  - Kubernetes YAML manifests for deploying the service.
- **examples/**  
  - **quickstart/**: A Kind cluster configuration to quickly set up a local Kubernetes environment.
- **Makefile**  
  - Provides common tasks for building, testing, and running the project.
- **Dockerfile**  
  - For containerizing the DNS gRPC server.

## Getting Started

### Prerequisites

- **Go** (1.x or later)
- **Docker** (for building container images)
- **Kubernetes** cluster (or [Kind](https://kind.sigs.k8s.io/) for local testing)
- **kubectl** (to interact with your cluster)
- **Make** (for build automation)

### Clone the Repository

```bash
git clone https://github.com/Networks-it-uc3m/l2sm-dns.git
cd l2sm-dns
```

### Update Module Path

If you fork or move the repository, update the module path in `go.mod`:
```bash
go mod edit -module github.com/<your-org>/l2sm-dns
go mod tidy
```

### Build and Test

Build the project using the Makefile:
```bash
make build
```

Run tests:
```bash
make test
```

### Running the gRPC Server

To run the DNS gRPC server locally, execute:
```bash
make run-server
```
The server listens on port `8081` and uses a Kubernetes configuration (in-cluster or via your local kubeconfig) to interact with the CoreDNS ConfigMap.

### Using the DNS Client

A sample DNS client is available in the `test/` directory. To send an `AddEntry` request, run:

```bash
go run test/client.go --test-add-entry --config ./test/config.yaml --pod your-pod --ip 10.0.1.2 --network your-network --scope global
```

This command will connect to the gRPC server, send the DNS entry details, and print the response.

### Deploying to Kubernetes

The repository includes Kubernetes manifests and kustomize configurations for a production-like deployment.

1. **RBAC Setup**  
   Apply RBAC resources:
   ```bash
   kubectl apply -k config/rbac/
   ```
2. **Deploy the Server**  
   Deploy the DNS server and its related resources:
   ```bash
   kubectl apply -k config/server/
   ```
3. **Quickstart with Kind (Optional)**  
   For a local cluster setup using Kind, run:
   ```bash
   kind create cluster --config examples/quickstart/kind-cluster.yaml
   kubectl apply -k config/server/
   ```

### CoreDNS Integration

The server uses the CoreDNSManager (see [pkg/coredns-manager/corednsmanager.go](pkg/coredns-manager/corednsmanager.go)) to update the CoreDNS `Corefile` dynamically. This allows you to add or remove DNS entries on the fly by modifying the CoreDNS ConfigMap (e.g., `l2smdns-coredns-config`).

## Makefile Targets

- **build**: Compiles the project.
- **test**: Runs unit and integration tests.
- **run-server**: Launches the gRPC DNS server.
- **docker-build**: Builds the Docker image.
- **docker-run**: Runs the Docker container locally.

## Reference to l2sm Repository

**l2sm-dns** is a key component of the broader [l2sm](https://github.com/Networks-it-uc3m/L2S-M) ecosystem. For further details, integration guidelines, and additional modules, please refer to the main [l2sm repository](https://github.com/Networks-it-uc3m/L2S-M).

## Contributing

Contributions are welcome! For guidelines on contributing, please check out the contribution instructions in the [l2sm repository](https://github.com/Networks-it-uc3m/L2S-M).

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.
