# captain-hook

a POC webhook relay that can be used to store and forward webhooks from GitHub to Jenkins.

`storing` - is currently done in memory with a backoff strategy, but the intention is that this can be moved to something external like a DB/K8S/something else...

## Installation

Should be installed in the same namespace as Jenkins.

```
helm repo add captain-hook https://jenkins-infra.github.io/captain-hook
helm install captain-hook captain-hook/captain-hook
```

## Configuration

Configuration options on the helm chart can be found [here](charts/captain-hook/README.md).

## Debugging

Once installed within a namespace, you can view the hooks within the system with:

```
kubectl get hooks
```

Or for more information:

```
kubectl get hooks -owide
```

