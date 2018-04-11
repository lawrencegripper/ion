set -x

namespace="$1"
if [[ -x $namespace ]];then
    $namespace = "default"
done


for j in $(kubectl get jobs --namespace=$namespace -o custom-columns=:.metadata.name)
do
    echo "deleting job '$j' from '$namespace' namespace"
    kubectl delete jobs $j --namespace $namespace &
done