# PULUMI

```bash
mkdir myproject && cd myproject && pulumi new kubernetes-go

pulumi config set isMinikube true

pulumi stack output ip

kubectl get svc
kubectl port-forward service/nginx-fe636420 8080:80

pulumi destroy
pulumy rm stack dev
```
