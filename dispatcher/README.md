# Dispatcher


# Testing

Integration tests expect the following environment variables

```env
AZURE_CLIENT_ID=
AZURE_CLIENT_SECRET=
AZURE_RESOURCE_GROUP=anExistingResourceGroupHere
AZURE_SUBSCRIPTION_ID=
AZURE_TENANT_ID=
AZURE_SERVICEBUS_NAMESPACE=anExistingNamespaceNameHere

```

The following vscode launch.json will kick them off. Edit the "program" property to point to your desired go package. 

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
