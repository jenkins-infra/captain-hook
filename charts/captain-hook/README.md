# captain-hook

A Helm chart for github.com/jenkins-infra/captain-hook

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

## Additional Information

This chart is best installed in the same namespace as Jenkins so that it can route webhooks directly
to the Jenkins service.

## Installing the Chart

To install the chart `captain-hook`:

```console
$ helm repo add captain-hook https://jenkins-infra.github.io/captain-hook
$ helm install captain-hook captain-hook/captain-hook
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| attemptRetryAfterInSeconds | int | `60` | Number of seconds the next retry should not be attempted before |
| autoscaling | object | `{"enabled":false,"maxReplicas":100,"minReplicas":1,"targetCPUUtilizationPercentage":80}` | Autoscaling configuration, disabled by default |
| forwardURL | string | `"http://jenkins:8080/github-webhook/"` | Url to send all webhook events to |
| fullnameOverride | string | `""` |  |
| hookPath | string | `"/hook"` | Path to listen for webhook events on |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"jenkinsciinfra/captain-hook"` |  |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `true` | Create an ingress resource for this service |
| ingress.hosts[0].paths[0].backend.service.name | string | `"captain-hook"` |  |
| ingress.hosts[0].paths[0].backend.service.port.number | int | `8080` |  |
| ingress.ingressClassName | string | `""` |  |
| ingress.tls | list | `[]` |  |
| insecureRelay | bool | `false` | Should we relay to insecure tls endpoints |
| maxAgeInSeconds | int | `3600` | Maximum age in seconds a successful webhook should be live for |
| maxAttempts | int | `10` | Maximum number of times this webhook should be attempted |
| nameOverride | string | `""` |  |
| replicaCount | int | `1` | Number of replicas to run |
| resources | object | `{}` |  |
| service.port | int | `8080` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
