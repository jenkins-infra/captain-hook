# captain-hook

A Helm chart for github.com/garethjevans/captain-hook

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

## Additional Information

This chart is best installed in the same namespace as Jenkins so that it can route webhooks directly
to the Jenkins service.

## Installing the Chart

To install the chart `captain-hook`:

```console
$ helm repo add captain-hook https://garethjevans.github.io/captain-hook
$ helm install captain-hook captain-hook/captain-hook
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| forwardURL | string | `"http://jenkins:8080/github-webhook/"` |  |
| fullnameOverride | string | `""` |  |
| hookPath | string | `"/hook"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"garethjevans/captain-hook"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `true` |  |
| ingress.hosts[0].paths[0].backend.service.name | string | `"captain-hook"` |  |
| ingress.hosts[0].paths[0].backend.service.port.number | int | `8080` |  |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| insecureRelay | bool | `false` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| service.port | int | `8080` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |
