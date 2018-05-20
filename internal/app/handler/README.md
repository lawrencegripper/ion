# Handler
The handler is co-scheduled with a module. The handler shares the execution environment with the module and is responsible for 2 things:

1. Fetching any input files and data from Ion's data plane.
2. Committing output files to Ion's data plane.

The handler runs as a separate container but leverages a shared volume with the module's container. When instructed to do so, the Handler will synchronize the data in this volume with Ion's data plane.

## Data Plane
The handler uses a provider for storing blob data and another provider for storing meta data. These backing stores are then responsible for handling the data persistence. These providers are know in the abstract as Ion's data plane.

![](../docs/ion-module.png)

## Getting the Handler
Now that you've Ionized your module, you probably want to check that it integrates with Ion's Handler properly. In order to do this, grab the latest release of the Handler binary from the [releases page](https://github.com/lawrencegripper/ion/releases).

## Running the Handler
The Handler binary can be ran as follows:

**Windows Powershell**
```powershell
.\handler.exe `
--context.correlationid=<correlationid> `
--context.eventid=<eventid>`
--context.parenteventid=<parenteventid> `
--context.name=<name> `
--loglevel=info `
--logfile=<logfile> `
--serverport=8080 `
--sharedsecret=<sharedsecret> `
--valideventtypes=<valideventtypes> `
--azureblobprovider=true `
--azureblobprovider.blobaccountkey=<blobaccountkey> `
--azureblobprovider.blobaccountname=<blobaccountname> `
--azureblobprovider.containername=<containername> `
--mongodbdocprovider=true `
--mongodbdocprovider.collection=<collection> `
--mongodbdocprovider.name=<name> `
--mongodbdocprovider.password=<password> `
--mongodbdocprovider.port=<port> `
--servicebuseventprovider=true `
--servicebuseventprovider.authorizationrulename=<authorizationrulename> `
--servicebuseventprovider.key=<key> `
--servicebuseventprovider.namespace=<namespace> `
--servicebuseventprovider.topic=<topic>
```

**Linux Bash**

```bash
./handler \
--context.correlationid=<correlationid> \
--context.eventid=<eventid> \
--context.parenteventid=<parenteventid> \
--context.name=<name> \
--loglevel=info \
--logfile=<logfile> \
--serverport=8080 \
--sharedsecret=<sharedsecret> \
--valideventtypes=<valideventtypes> \
--azureblobprovider=true \
--azureblobprovider.blobaccountkey=<blobaccountkey> \
--azureblobprovider.blobaccountname=<blobaccountname> \
--azureblobprovider.containername=<containername> \
--mongodbdocprovider=true \
--mongodbdocprovider.collection=<collection> \
--mongodbdocprovider.name=<name> \
--mongodbdocprovider.password=<password> \
--mongodbdocprovider.port=10255 \
--servicebuseventprovider=true \
--servicebuseventprovider.authorizationrulename=<authorizationrulename> \
--servicebuseventprovider.key=<key> \
--servicebuseventprovider.namespace=<namespace> \
--servicebuseventprovider.topic=<topic>
```

> **NOTE:** You can supply the `--development=true` argument to enabled development mode

### Development Mode
Development mode allows you to run the handler without the Dispatcher. This will leverage the filesystem and in-memory providers to handle blobs, metadata and events.

Development mode will also dump data of interest out to a `dev` folder to help you debug issues.

> **Coming Soon:** Offline Mode - use a local Dispatcher to simulate Ion without any external services.

# Data Handling
A module should work with files on the local file system as any other process would.
However, _persistent_ data is expected to be written/read using a particular directory structure:

## `in/data`
Any input files that your module needs will be available in the input blob directory `in/data` after you have called [ready](#ready).

## `in/eventmeta.json`
Any input values that your module needs will be available in the file `in/eventmeta.json`.
This file will need to be deserialized from JSON into an instance of `common.KeyValuePairs`.

## `out/data`
Any output files you wish to store should be written to `out/data`. These will be persisted after you call [done](#done).

## `out/insights.json`
Any insights you wish to export should be written to the JSON file `out/insights.json`. This is intended for data you want to store for later analysis and does not get passed to subsequent modules.

### `Insight Schema`

Key value pairs can only currently be stored as strings and should be serialized from the type `common.KeyValuePairs`
```json
[
  {
  "key": "key",
  "value": "value"
  },
  {
    "key": "key1",
    "value": "value1"
  },
  ...
]
```

## `out/events`
Any events you wish to publish should be stored as JSON files in `out/events`. Any _optional_ key/value data will be made available to subsequent modules in their `in/meta.json` file. _Required_ key/value data will be extracted.

### `Events Schema`

Key value pairs can only currently be stored as strings.
```json
[
  {
    "key": "eventType", [required]
    "value": "myEventType"
  },
  {
    "key": "files", [required]
    "value": "myblob.png,myblob2.png"
  },
  {
    "key": "key", [optional]
    "value": "value"
  },
  ...
]
```

## Temporary Files
Any temporary files you wish to use can be written into any other directory in the file system i.e. `/tmp`. These files will be lost when the Job is complete.