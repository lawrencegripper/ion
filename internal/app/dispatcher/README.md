# Dispatcher
A Dispatcher is responsible for picking up events from a messaging topic and then scheduling a job using an appropriate provider i.e. Kubernetes. Once the Dispatcher has scheduled the job, it will monitor its progress until termination. Once the job is terminated, depending on its exit status, the Dispatcher will either mark the event as fulfilled or not. If an event is marked as not fulfilled or the job times out, the event will be requeued and re-processed. If this continues multiple times, the event will eventually end up being put on a dead letter queue.

![](../docs/dispatcher.png)

# Getting the Dispatcher
When you're ready to dispatch your module onto an execution environment, you'll need to download the Dispatcher from the [releases page](https://github.com/lawrencegripper/ion/releases).

# Running the Dispatcher
You can run the Dispatcher locally against a Kubernetes cluster as long as you have a Kubernetes config set. Otherwise, you'll need to deploy the Dispatcher to Kubernetes so it can use the built in config.

## Running the Dispatcher locally
Once you have the Dispatcher binary, you can simply run it using one of the following commands:

**Windows Powershell**
```powershell
.\dispatcher `
--handler.azureblobprovider.blobaccountkey=<blobaccountkey> `
--handler.azureblobprovider.blobaccountname=<blobaccountname> `
--loglevel=debug `
--modulename=<modulename> `
--kubernetesnamespace=default `
--handler.mongodbdocprovider=true `
--handler.mongodbdocprovider.collection=<collection> `
--handler.mongodbdocprovider.name=<name> `
--handler.mongodbdocprovider.password=<password> `
--handler.mongodbdocprovider.port=10255 `
--handler.serverport=8080 `
--servicebusnamespace=<servicebusnamespace> `
--resourcegroup=<resourcegroup> `
--subscribestoevent=<subscribestoevent> `
--eventspublished=<eventspublished> `
--job.workerimage=<workerimage> `
--job.handlerimage=<handlerimage> `
--job.retrycount=0 `
--clientid=<clientid> `
--clientsecret=<clientsecret> `
--tenantid=<tenantid> `
--subscriptionid=<subscriptionid>
```

**Linux Bash**
```bash
./dispatcher \
--handler.azureblobprovider.blobaccountkey=<blobaccountkey> \
--handler.azureblobprovider.blobaccountname=<blobaccountname> \
--loglevel=debug \
--modulename=<modulename> \
--kubernetesnamespace=default \
--handler.mongodbdocprovider=true \
--handler.mongodbdocprovider.collection=<collection> \
--handler.mongodbdocprovider.name=<name> \
--handler.mongodbdocprovider.password=<password> \
--handler.mongodbdocprovider.port=10255 \
--handler.serverport=8080 \
--servicebusnamespace=<servicebusnamespace> \
--resourcegroup=<resourcegroup> \
--subscribestoevent=<subscribestoevent> \
--eventspublished=<eventspublished> \
--job.workerimage=<workerimage> \
--job.handlerimage=<handlerimage> \
--job.retrycount=0 \
--clientid=<clientid> \
--clientsecret=<clientsecret> \
--tenantid=<tenantid> \
--subscriptionid=<subscriptionid>
```

## Running the Dispatcher on Kubernetes
We've provided a [Helm](https://helm.sh/) chart to make it easy to deploy the Dispatcher to Kubernetes.

> NOTE: Ensure you have helm installed on the Kubernetes cluster. Instructions on doing so are [here](https://docs.helm.sh/using_helm/#installing-helm).

Simply update the `values.yaml` file inside the `dispatcher/helm/` directory with your desired configuration and then use
helm to install the chart.

`helm install -f values.yaml ./dispatcher`

# Testing the Dispatcher

Integration tests expect the following environment variables

```env
AZURE_CLIENT_ID=
AZURE_CLIENT_SECRET=
AZURE_RESOURCE_GROUP=anExistingResourceGroupHere
AZURE_SUBSCRIPTION_ID=
AZURE_TENANT_ID=
AZURE_SERVICEBUS_NAMESPACE=anExistingNamespaceNameHere

```
Write the above environment variables to a file such as `private.env` in the workspace root.
Then you can use the following vscode `launch.json` to kick the tests off. Edit the "program" property to point to your desired go package. 

```json
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "remotePath": "",
            "port": 2345,
            "host": "127.0.0.1",
            "program": "${workspaceRoot}/servicebus",
            "envFile": "${workspaceRoot}/.vscode/private.env",
            "args": [
                "-test.v",  
                "-test.timeout",
                "5m"
            ],
            "showLog": true
        }
    ]
}
```
