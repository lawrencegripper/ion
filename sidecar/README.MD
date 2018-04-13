# Sidecar
The sidecar is deployed alongside your module. The sidecar is responsible for 2 things:

1. Setting up the execution environment
2. Synchronizing state to external providers

# Data
A module should work with files on the local file system with as little knowledge of the surrounding system as possible.
However, data is expected to be written/read using a particular directory structure:

## `in/data`
Any input files that your module needs will be available in the input blob directory `in/data` after you have called [ready](#ready).

## `in/meta.json`
Any input values that your module needs will be available in the file `in/meta.json` after you have called [ready](#ready).
This file will need to be deserialized from JSON into an instance of `common.KeyValuePairs`.

## `out/data`
Any output files you wish to store should be written to `out/data`. These will be persisted after you call [done](#done).

## `out/meta.json`
Any insights you wish to store should be written to the JSON file `out/meta.json`. This will be persisted after you call [done](#done).

### `Insight Schema`

Key value pairs can only currently be stored as strings and should be serialized from the type `common.KeyValuePairs`
```json
[
  {
  "key": "key `
  "value": "value"
  },
  {
    "key": "key1 `
    "value": "value1"
  },
  ...
]
```

## `out/events`
Any events you wish to publish should be stored as JSON files in `out/events`

### `Events Schema`

Key value pairs can only currently be stored as strings.
```json
[
  {
    "key": "eventType ` [required]
    "value": "myEventType"
  },
  {
    "key": "files ` [required]
    "value": "myblob.png,myblob2.png"
  },
  {
    "key": "key ` [optional]
    "value": "value"
  },
  ...
]
```

## Temporary Files
Any temporary files you wish to use can be written into any other directory in the file system i.e. `/tmp`. These files will be lost when the Job is complete.

# API
Calls between the module and the sidecar are authenticated using a shared secret.
This is not currently used to encrypt communication between the 2 parties and is only used as a simple check.
The shared secret will be injected into your module as an environment variable: `SHARED_SECRET`.
Please add this as a header to all HTTP requests sent to the sidecar.

```python
shared_secret = os.environ["SHARED_SECRET"]
headers = {"secret": shared_secret}
res = requests.get(parent_meta_url, headers=headers)
```

## Ready
A call to `ready` is used by the module to inform the sidecar to prepare the execution environment i.e. pulling any input files, creating directories, etc.
The call will block whilst the environment is setup, once successfully returned, the module can start processing files.

> **NOTE:** This call should be made in a retry loop as process execution order is not guaranteed by the execution environment. The example module demonstrates this.

> **NOTE:** _Only_ make this call once.

## `GET /ready`
Synchronous call to setup the execution environment when module ready. Only make this call once.

**headers:**

`{"secret": "<shared_secret>"}`

## Done
A call to `done` is used by the module to inform the sidecar to commit the current state to external providers.
This includes 
The call will block whilst the environment is setup, once successfully returned, the module can start processing files.

> **NOTE:** This call should be made in a retry loop as process execution order is not guaranteed by the execution environment. The example module demonstrates this.

> **NOTE:** _Only_ make this call once.

## `GET /done`
Synchronous call to commit state when module is done. 

**headers:**

`{"secret": "<shared_secret>"}`

## Getting the Sidecar
Now that you've Ionized your module, you probably want to check that it integrates with Ion's Sidecar properly. In order to do this, grab the latest release of the Sidecar binary from the [releases page](https://github.com/lawrencegripper/ion/releases).

## Running the Sidecar
The Sidecar binary can be ran as follows:

**Windows Powershell**
.\sidecar.exe --context.correlationid=<correlationid>`
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
--mongodbmetaprovider=true `
--mongodbmetaprovider.collection=<collection> `
--mongodbmetaprovider.name=<name> `
--mongodbmetaprovider.password=<password> `
--mongodbmetaprovider.port=<port> `
--servicebuseventprovider=true `
--servicebuseventprovider.authorizationrulename=<authorizationrulename> `
--servicebuseventprovider.key=<key> `
--servicebuseventprovider.namespace=<namespace> `
--servicebuseventprovider.topic=<topic>

**Linux Bash**
./sidecar --context.correlationid=<correlationid> \
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
--mongodbmetaprovider=true \
--mongodbmetaprovider.collection=<collection> \
--mongodbmetaprovider.name=<name> \
--mongodbmetaprovider.password=<password> \
--mongodbmetaprovider.port=10255 \
--servicebuseventprovider=true \
--servicebuseventprovider.authorizationrulename=<authorizationrulename> \
--servicebuseventprovider.key=<key> \
--servicebuseventprovider.namespace=<namespace> \
--servicebuseventprovider.topic=<topic>

> **NOTE:** You can supply the `--development=true` argument to enabled development mode

### Development Mode
Development mode allows you to run the sidecar without the dispatcher. This will leverage the filesystem and in-memory providers to handle blobs, metadata and events.

Development mode will also dump data of interest out to the `out/dev` folder to help you debug issues.

> **Coming Soon:** Offline Mode - Develop with no external dependencies
