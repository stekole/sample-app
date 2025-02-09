# Helm Chart

This helm chart is great! It give the ability to loop over configuration and template Kubernetes resources. Making deployments standardized and still flexible!

## Chart Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `deployments` | Map of deployment configurations | `{}` |
| `deployments.<name>.image.repository` | Container image repository | Required |
| `deployments.<name>.image.tag` | Container image tag | Required |
| `deployments.<name>.image.pullPolicy` | Image pull policy | Required |  
| `deployments.<name>.appPort` | Application port | `8080` |
| `deployments.<name>.appPortName` | Name of the application port | `http` |
| `deployments.<name>.metricsEnabled` | Enable Prometheus metrics scraping | `false` |
| `deployments.<name>.metricsPort` | Port for metrics scraping | `8080` |
| `deployments.<name>.metricsPath` | Path for metrics scraping | `/metrics` |
| `deployments.<name>.resources` | Container resource requests/limits | `{limits: {cpu: 100m, memory: 128Mi}, requests: {cpu: 100m, memory: 128Mi}}` |
| `deployments.<name>.livenessProbe` | Container liveness probe configuration | `{}` |
| `deployments.<name>.readinessProbe` | Container readiness probe configuration | `{}` |
| `deployments.<name>.startupProbe` | Container startup probe configuration | `{}` |
| `deployments.<name>.env` | Environment variables for the container | `[]` |
| `deployments.<name>.annotations` | Additional annotations for deployment/service/pod | `{}` |
| `deployments.<name>.service.appPortName` | Service port name | `http` |
| `deployments.<name>.service.protocol` | Service port protocol | `TCP` |
| `deployments.<name>.service.port` | Service port number | Same as appPort |
| `deployments.<name>.service.targetPort` | Service target port number | Same as appPort |

## Addons

Adding other charts can be done by adding the following to the `Chart.yaml` dependancies:

```yaml
dependencies:
  - name: <chart-name>
    version: <chart-version>
    repository: <chart-repository>
```

look in the `Chart.yaml` file for an example for Redis. 

## Helpers

Helm helper files are a great way to add common functionality to your charts. For example, a common set of labels are usually required for all resources. Including the helper allows to write the values once, but inckude it in multiple locations. Review the helper.tpl for some examples.

## Globals

These are fun and awesome. Very helpful in a distributed environment. Currently I do not have an example in here.
