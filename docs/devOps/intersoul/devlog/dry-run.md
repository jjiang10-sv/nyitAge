The `kubectl` command provides a `--dry-run` option that allows you to preview the results of an operation without actually applying changes to the cluster. This is useful when you want to test or see what changes will be made before executing them.

### Types of `--dry-run`

There are two modes for `--dry-run`:

- **`client` (formerly `--dry-run=client`)**: Runs the dry run entirely on the client side, showing the output without making any changes to the cluster. This is useful when you just want to see what would be sent to the Kubernetes API server.
  
- **`server` (formerly `--dry-run=server`)**: Sends the request to the server and checks whether it would succeed without actually applying the change. This is useful to ensure that the resource would be accepted by the API server, even though it isn't actually created or modified.

Starting with Kubernetes 1.18, `--dry-run=client` became the default behavior if you just use `--dry-run`.

### Basic Usage Examples:

#### 1. **Creating resources with `--dry-run`**
You can use `--dry-run=client` or `--dry-run=server` with resource creation commands like `kubectl create` or `kubectl apply`.

- **Create a deployment dry-run (client-side)**:
  ```bash
  kubectl create deployment my-deployment --image=nginx --dry-run=client -o yaml
  ```
  This will print the YAML output of the deployment without actually creating it.

- **Create a pod dry-run (server-side)**:
  ```bash
  kubectl run my-pod --image=nginx --dry-run=server -o yaml
  ```
  This will send the request to the API server to validate the pod configuration, but no pod will be created.

#### 2. **Applying a manifest with `--dry-run`**
If you want to apply changes from a manifest file but see what will happen first:

- **Client-side dry-run**:
  ```bash
  kubectl apply -f my-manifest.yaml --dry-run=client
  ```

- **Server-side dry-run**:
  ```bash
  kubectl apply -f my-manifest.yaml --dry-run=server
  ```

#### 3. **Patching resources with `--dry-run`**
When modifying existing resources using `kubectl patch`, you can use the `--dry-run` flag to test the patch:

```bash
kubectl patch deployment my-deployment -p '{"spec": {"replicas": 2}}' --dry-run=client
```

#### 4. **Deleting resources with `--dry-run`**
If you want to test deleting a resource without actually doing it, you can use:

```bash
kubectl delete pod my-pod --dry-run=client
```

This will display the output indicating that the pod would be deleted, but it won’t actually delete the pod.

### Examples for `kubectl` dry-run use cases:

#### Create a service and see YAML without applying:
```bash
kubectl expose deployment my-deployment --type=LoadBalancer --name=my-service --dry-run=client -o yaml
```

#### Scale a deployment (client-side) without actually scaling:
```bash
kubectl scale deployment my-deployment --replicas=3 --dry-run=client
```

### Conclusion:

The `--dry-run` option is an excellent way to validate and preview changes before making them to a live cluster, allowing you to debug and avoid issues by ensuring commands and resources are correct before actual deployment.

In Helm, the `--dry-run` flag allows you to simulate the deployment or upgrade of a chart without actually making any changes to the cluster. It shows what would happen if the command were executed, helping you to check the generated Kubernetes manifest and understand the changes that would occur. 

### Usage of `--dry-run` in Helm:

#### 1. **Install a Helm Chart with `--dry-run`**

You can use the `--dry-run` flag during the installation of a chart to preview what will be applied without actually creating any resources in the cluster.

```bash
helm install my-release my-chart --dry-run --debug
```

- `my-release`: The release name.
- `my-chart`: The chart name or path.
- `--dry-run`: Simulates the install without deploying anything.
- `--debug`: Shows additional debug information, including the rendered templates.

This command will display the rendered Kubernetes YAML manifests that would be created, without actually installing the release.

#### 2. **Upgrade a Release with `--dry-run`**

Similarly, you can use `--dry-run` during an upgrade to see what changes will be applied without making actual updates to the release:

```bash
helm upgrade my-release my-chart --dry-run --debug
```

This shows how the release would be upgraded, giving you the chance to verify the changes before applying them.

#### 3. **Example Output**

When you run `helm install` or `helm upgrade` with `--dry-run`, Helm will:
- Render the Kubernetes manifests (YAML files).
- Simulate the installation or upgrade.
- Print the rendered manifests to the console, showing what will happen without applying the changes.

Example output:
```bash
NAME: my-release
LAST DEPLOYED: Thu Oct 17 11:22:00 2024
NAMESPACE: default
STATUS: pending-install
REVISION: 1
TEST SUITE: None
HOOKS:
---
# Source: my-chart/templates/hook.yaml
apiVersion: v1
kind: Pod
...

MANIFEST:
---
# Source: my-chart/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
...
```

#### 4. **Additional Options with `--dry-run`**

You can combine the `--dry-run` flag with other options to control the output, for example:

- **Specifying a namespace**:
  ```bash
  helm install my-release my-chart --namespace my-namespace --dry-run
  ```

- **Using values files**:
  ```bash
  helm install my-release my-chart --values my-values.yaml --dry-run
  ```

- **Simulating uninstalling a release**:
  ```bash
  helm uninstall my-release --dry-run
  ```

### Why Use `--dry-run`?

- **Preview changes**: Before making actual changes to your cluster, `--dry-run` allows you to verify that everything looks correct, such as ensuring values and templates render as expected.
- **Debugging**: With `--debug` and `--dry-run`, you can inspect the detailed output and debug any issues with the Helm templates or values before applying them.
- **Validate manifests**: Helps confirm that your manifests are properly formatted and will be accepted by the Kubernetes API.

### Summary:
Using `--dry-run` in Helm provides a safe and effective way to simulate installs, upgrades, and other Helm operations without changing the cluster, ensuring that everything works as expected before applying changes.