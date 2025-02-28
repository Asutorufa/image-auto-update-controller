#

auto update images that pods have `asutorufa.github.io/image-updater=true` label

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    asutorufa.github.io/image-updater: "true"
```

__Current Only Support Single Node__  
For support Multiple node, we need impl distributed lock and handless service.  

## install

```shell
cd infra
terraform init
terraform apply -var="cri-socket=/run/k0s/containerd.sock"
```
