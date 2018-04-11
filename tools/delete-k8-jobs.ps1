param(
    $namespace="default"
)

$jobs=$(kubectl get jobs --namespace=$namespace -o custom-columns=:.metadata.name)
For ($i = 0; $i -lt $jobs.Length; $i++) {
    $job = $jobs[$i]
    If ($job -ne "") {
        Write-Host "deleting job '$job' from '$namespace' namespace"
        kubectl delete jobs $job --namespace $namespace
    }
}