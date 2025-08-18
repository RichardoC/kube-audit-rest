
testing/setup.sh

# Force auditing to stdout
kubectl -n kube-audit-rest patch  deployment kube-audit-rest --patch='{"spec":{"template":{"spec":{"$setElementOrder/containers":[{"name":"kube-audit-rest"}],"containers":[{"args":["--audit-to-std-log"],"name":"kube-audit-rest"}]}}}}'

kubectl delete ns hacker --ignore-not-found=true
kubectl create ns hacker


kubectl -n hacker create secret generic hacking-creds --from-literal="DB_PASSWORD"="V3rySecr3t"

kubectl -n hacker run monero-miner --force=true --image alpine -- tail -f /dev/null

kubectl -n kube-audit-rest logs deployment/kube-audit-rest | grep zapio| grep hacker | jq '.msg | fromjson '
