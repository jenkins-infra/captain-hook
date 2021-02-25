# captain-hook

a POC webhook relay that can be used to store and forward webhooks from GitHub to Jenkins.

`storing` - is currently done in memory with a backoff strategy, but the intention is that this can be moved to something external like a DB/K8S/something else...

## Installation

Should be installed in the same namespace as Jenkins.

```
helm repo add https://garethjevans.github.io/captain-hook
helm install captain-hook captain-hook/captain-hook
```

## Configuration

TODO


