# Handler
The handler can be run in 2 discrete modes:
* Prepare
* Commit

## Prepare
When the handler is run in `prepare` mode, it will create the required directory structure for the module to use and populate it with any data passed in from a previous module.

## Commit
When then handler is run in `commit` mode, it will take any data written out by the module during execution and persist it into Ion's data plane.

## Run Order
Ion will execute Jobs in a specific order:
1. The handler in prepare mode
2. The module
3. The handler in commit mode

## Data Plane
Ion's data plane provides 3 main capabilities:
* Blob storage
* Document storage
* Event publishing
These 3 capabilities are fulfilled by providers. The data plane relies on these providers to interface with whatever service is backing their interface. For instance there is an [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/) provider that is responsible for storing and retrieving data from the Azure Blob Storage service. This means that you service can only operate at the service level defined by your data plane providers. If you choose to use a single instance of MongoDB as your document storage provider and that instance is deleted, then you will lose your data.

![](../../../docs/ion3.PNG)

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

Development mode will also dump data of interest out to a `.dev` folder to help you debug issues.

> **Coming Soon:** Offline Mode - use a local Dispatcher to simulate Ion without any external services.

# Data Handling
A module should work with files on the local file system as any other process would.
However, _persistent_ data is expected to be written/read using a particular directory structure:

## `/ion/in/data`
Any input files that your module needs will be available in the input blob directory `/ion/in/data`.

## `/ion/in/eventmeta.json`
Any input values that your module needs will be available in the file `/ion/in/eventmeta.json`.
This file will need to be deserialized from JSON into an instance of `common.KeyValuePairs`.

## `/ion/out/data`
Any output files you wish to store should be written to `/ion/out/data`.

## `/ion/out/insights.json`
Any insights you wish to export should be written to the JSON file `/ion/out/insights.json`. This is intended for data you want to store for later analysis and does not get passed to subsequent modules.

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

## `/ion/out/events`
Any events you wish to publish should be stored as JSON files in `/ion/out/events`. Any _optional_ key/value data will be made available to subsequent modules in their `/ion/in/eventmeta.json` file. _Required_ key/value data will be extracted.

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