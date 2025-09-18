# Ginkgo-Based Conformance Tester

This document describes how to run, configure, and debug the Ginkgo-based conformance tests for Kubermatic. This new testing framework replaces the previous conformance tester with a more flexible, maintainable, and powerful solution built on Ginkgo and Gomega.

## Benefits of the Ginkgo-Based Approach

The new test suite offers several advantages over the previous implementation:

- **Declarative and Readable Tests**: By using Ginkgo's BDD (Behavior-Driven Development) style, tests are structured with `Describe`, `Context`, and `It` blocks, making them easier to read and understand.

- **Dynamic Test Generation**: The test suite automatically discovers providers by inspecting the `kubermaticv1.DatacenterSpec` struct. This means new providers are included in tests without needing to modify the test code.

- **Data-Driven Scenarios**: Tests for each provider are built from a list of "user stories" defined in `TestSettings`. This makes it simple to add new test variations and configurations (e.g., for KubeVirt, testing different instance types or features) in a structured way.

- **Flexible and Selective Test Execution**: You have fine-grained control over which tests to run:
  - Run all tests for a specific provider.
  - Run only a subset of user stories using the `testSettings` configuration option.
  - Leverage Ginkgo's powerful `--focus` and `--skip` flags to run or ignore specific tests on the fly.

- **Powerful Tooling**: It leverages the rich ecosystem of Ginkgo, including parallel test execution (`ginkgo -p`), advanced reporting (JUnit, JSON), and a mature debugging experience.

## Installation

Before you can run the tests, you need to install the `ginkgo` CLI:

```bash
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

## Running Tests

Tests can be run using either `go test` or the `ginkgo` CLI. Before running, ensure the `CONFORMANCE_TESTER_CONFIG_FILE` environment variable points to your configuration file.

```bash
export CONFORMANCE_TESTER_CONFIG_FILE=./config.yaml
```

### Using `go test`

```bash
go test -v ./pkg/ginkgo/...
```

### Using the `ginkgo` CLI

The `ginkgo` CLI is the recommended way to run tests, as it provides more control.

```bash
# Run all tests verbosely
ginkgo -v ./pkg/ginkgo/...

# Run tests in parallel to speed up execution
ginkgo -p -v ./pkg/ginkgo/...

# Focus on a specific provider (e.g., KubeVirt)
ginkgo -v --focus="KubeVirt" ./pkg/ginkgo/...

# Focus on a specific user story
ginkgo -v --focus="with a secondary disk" ./pkg/ginkgo/...
```

## Configuration

The tests are configured using a YAML file specified by the `CONFORMANCE_TESTER_CONFIG_FILE` environment variable.

### Example Configuration

```yaml
# A prefix for all created resources.
namePrefix: "ginkgo-test"

# A list of providers to test.
providers:
- KubeVirt

# A list of Kubernetes releases to test.
releases:
- "1.25"

# A list of enabled operating system distributions.
enableDistributions:
- ubuntu

# A list of specific test settings (user stories) to run. If empty or omitted,
# all tests for the selected providers are run. The descriptions must match exactly.
testSettings:
  - "with a specific instancetype"
  - "with a secondary disk"

# The file to write the test results to.
resultsFile: "results.json"

# If set, only scenarios that failed in a previous run will be executed.
retryFailedScenarios: false

# Cluster settings
deleteClusterAfterTests: true
nodeCount: 1

# Paths
reportsRoot: "_reports"
logDirectory: "_logs"

# Kubermatic settings
kubermaticNamespace: "kubermatic"
kubermaticSeedName: "kubermatic"

# Secrets for the providers
secrets:
  kubevirt:
    kubeconfig: "/path/to/your/kubevirt-kubeconfig"
    kkpDatacenter: "kubevirt-dc"
```

## Secrets Management

Provider secrets can be provided directly in the configuration file or loaded from external files. To load a secret from a file, append `File` to the secret key and provide a path to the file.

**Example:**

```yaml
secrets:
  kubevirt:
    # Provide the Kubeconfig directly
    kubeconfig: "apiVersion: v1..."
  hetzner:
    # Load the token from a file
    tokenFile: "/path/to/hetzner-token"
```

## Test Reporting

The tests generate reports in JUnit XML format. By default, these reports are saved in the `_reports` directory. You can change this location using the `reportsRoot` key in your configuration file. The name of the file will be `junit_ginkgo.xml`.

When using the `ginkgo` CLI, you can also generate other types of reports:

```bash
# Generate a JSON report
ginkgo --json-report=report.json ./pkg/ginkgo/...

# Generate a TeamCity report
ginkgo --teamcity-report=report.teamcity ./pkg/ginkgo/...
```

## Debugging

To debug the tests, you can use the Delve debugger.

1.  Set the `CONFORMANCE_TESTER_CONFIG_FILE` environment variable.

    ```bash
    export CONFORMANCE_TESTER_CONFIG_FILE=./config.yaml
    ```

2.  Build the test binary:

    ```bash
    go test -c ./pkg/ginkgo/... -o ginkgo.test
    ```

3.  Run the test binary with Delve:

    ```bash
    dlv exec ./ginkgo.test -- -test.v -ginkgo.v
    ```

You can then set breakpoints and inspect the state of the application as you would with any other Go program.

## Advanced CLI Usage

While using a YAML configuration file is recommended for most scenarios, it is also possible to run the tests by providing all options as command-line flags. This is useful for quick runs or for integration into scripts where creating a config file might be cumbersome.

### Example: Running KubeVirt Tests via CLI

Here is a complex example that runs specific KubeVirt tests for a single Kubernetes release without a `config.yaml` file. Note that this requires passing provider-specific secrets directly on the command line.

```bash
# Unset the config file variable to ensure CLI flags are used
unset CONFORMANCE_TESTER_CONFIG_FILE

# Run the tests using ginkgo and pass all parameters as flags
ginkgo -v ./pkg/ginkgo/... -- \
  --name-prefix="ginkgo-cli" \
  --providers="kubevirt" \
  --releases="1.25" \
  --enable-distributions="ubuntu" \
  --test-settings="with a specific instancetype,with a secondary disk" \
  --delete-cluster-after-tests=true \
  --node-count=1 \
  --kubermatic-seed-name="kubermatic" \
  --kubevirt-kkp-datacenter="kubevirt-dc" \
  --kubevirt-kubeconfig="/path/to/your/kubevirt-kubeconfig"
```

**Explanation of Flags:**

The following is a comprehensive list of flags that can be used to configure the test runner from the command line.

-   `--name-prefix`: A prefix for all created resources (e.g., clusters).
-   `--providers`: A comma-separated list of cloud providers to test (e.g., `aws`, `kubevirt`).
-   `--releases`: A comma-separated list of Kubernetes versions to test (e.g., `1.25`, `1.26`).
-   `--enable-distributions`: A comma-separated list of operating systems to test (e.g., `ubuntu`, `centos`).
-   `--test-settings`: A comma-separated list of exact test descriptions to run. This allows for fine-grained control over which user stories are executed.
-   `--delete-cluster-after-tests`: A boolean (`true` or `false`) that controls whether the user cluster is deleted after tests complete.
-   `--node-count`: The number of worker nodes to create for the user cluster.
-   `--kubermatic-seed-name`: The name of the seed cluster where the test cluster will be created.
-   `--kubermatic-namespace`: The namespace where Kubermatic is installed.
-   `--kubermatic-project`: The Kubermatic project to use. If left empty, a new one is created.
-   `--client`: Controls how to interact with KKP; can be either `api` or `kube`.
-   `--existing-cluster-label`: If specified, tests will run against an existing cluster matching this label instead of creating a new one.
-   `--exclude-distributions`: A comma-separated list of distributions to exclude from testing.
-   `--tests`: A comma-separated list of specific tests to run.
-   `--exclude-tests`: A comma-separated list of tests to exclude.
-   `--scenario-options`: A comma-separated list of additional options for test scenarios.
-   `--repo-root`: The root path for Kubernetes repositories.
-   `--kubermatic-parallel-clusters`: The number of clusters to test in parallel.
-   `--reports-root`: The root directory for test reports.
-   `--log-directory`: The directory where container logs will be saved.
-   `--kubermatic-cluster-timeout`: The timeout for cluster creation.
-   `--node-ready-timeout`: The timeout for nodes to become ready.
-   `--custom-test-timeout`: A custom timeout for specific tests like PVC/LB.
-   `--user-cluster-poll-interval`: The interval for polling user cluster conditions.
-   `--wait-for-cluster-deletion`: A boolean that determines whether to wait for cluster deletion to complete.
-   `--node-ssh-pub-key`: The path to an SSH public key to deploy on each node.
-   `--enable-dualstack`: A boolean to enable dual-stack (IPv4+IPv6) networking.
-   `--enable-konnectivity`: A boolean to enable Konnectivity instead of OpenVPN.
-   `--update-cluster`: If `true`, the cluster will be updated to the next minor release and tests will be run again.
-   `--results-file`: The path to a JSON file for saving test results.
-   `--retry`: If `true`, only failed scenarios from a previous run (indicated by `--results-file`) will be executed.

**Provider-Specific Secret Flags:**

Each provider has its own set of flags for secrets and configuration. Below is a comprehensive list.

-   **Anexia**
    -   `--anexia-token`: Anexia API Token
    -   `--anexia-template-id`: The template ID to use for nodes.
    -   `--anexia-vlan-id`: The VLAN ID.
    -   `--anexia-kkp-datacenter`: The KKP datacenter to use.
-   **AWS**
    -   `--aws-access-key-id`: AWS Access Key ID.
    -   `--aws-secret-access-key`: AWS Secret Access Key.
    -   `--aws-kkp-datacenter`: The KKP datacenter to use.
-   **Azure**
    -   `--azure-client-id`: Azure Client ID.
    -   `--azure-client-secret`: Azure Client Secret.
    -   `--azure-tenant-id`: Azure Tenant ID.
    -   `--azure-subscription-id`: Azure Subscription ID.
    -   `--azure-kkp-datacenter`: The KKP datacenter to use.
-   **DigitalOcean**
    -   `--digitalocean-token`: DigitalOcean API Token.
    -   `--digitalocean-kkp-datacenter`: The KKP datacenter to use.
-   **GCP**
    -   `--gcp-service-account`: GCP Service Account JSON.
    -   `--gcp-network`: The network to use.
    -   `--gcp-subnetwork`: The subnetwork to use.
    -   `--gcp-kkp-datacenter`: The KKP datacenter to use.
-   **Hetzner**
    -   `--hetzner-token`: Hetzner API Token.
    -   `--hetzner-kkp-datacenter`: The KKP datacenter to use.
-   **KubeVirt**
    -   `--kubevirt-kubeconfig`: Path to the Kubeconfig for the KubeVirt cluster.
    -   `--kubevirt-kkp-datacenter`: The KKP datacenter to use.
-   **OpenStack**
    -   `--openstack-domain`: OpenStack domain.
    -   `--openstack-project`: OpenStack project.
    -   `--openstack-project-id`: OpenStack project ID.
    -   `--openstack-username`: OpenStack username.
    -   `--openstack-password`: OpenStack password.
    -   `--openstack-kkp-datacenter`: The KKP datacenter to use.
-   **VSphere**
    -   `--vsphere-username`: vSphere username.
    -   `--vsphere-password`: vSphere password.
    -   `--vsphere-kkp-datacenter`: The KKP datacenter to use.
-   **Alibaba**
    -   `--alibaba-access-key-id`: Alibaba Access Key ID.
    -   `--alibaba-access-key-secret`: Alibaba Access Key Secret.
    -   `--alibaba-kkp-datacenter`: The KKP datacenter to use.
-   **Nutanix**
    -   `--nutanix-username`: Nutanix username.
    -   `--nutanix-password`: Nutanix password.
    -   `--nutanix-csi-username`: Nutanix CSI Prism Element username.
    -   `--nutanix-csi-password`: Nutanix CSI Prism Element password.
    -   `--nutanix-csi-endpoint`: Nutanix CSI Prism Element endpoint.
    -   `--nutanix-proxy-url`: HTTP Proxy URL to access the endpoint.
    -   `--nutanix-cluster-name`: The Nutanix cluster name.
    -   `--nutanix-project-name`: The Nutanix project name.
    -   `--nutanix-subnet-name`: The Nutanix subnet name.
    -   `--nutanix-kkp-datacenter`: The KKP datacenter to use.
-   **VMware Cloud Director**
    -   `--vmware-cloud-director-username`: VMware Cloud Director username.
    -   `--vmware-cloud-director-password`: VMware Cloud Director password.
    -   `--vmware-cloud-director-organization`: VMware Cloud Director organization.
    -   `--vmware-cloud-director-vdc`: VMware Cloud Director Organizational VDC.
    -   `--vmware-cloud-director-ovdc-networks`: Comma-separated list of OVDC network names.
    -   `--vmware-cloud-director-kkp-datacenter`: The KKP datacenter to use.
-   **RHEL**
    -   `--rhel-subscription-user`: Red Hat Enterprise subscription user.
    -   `--rhel-subscription-password`: Red Hat Enterprise subscription password.
    -   `--rhel-offline-token`: Red Hat Enterprise offline token.
