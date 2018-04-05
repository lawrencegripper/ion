# Sidecar
The sidecar is deployed alongside your module. The sidecar is responsible for 2 things:

1. Setting up the execution environment
2. Synchronizing state to external providers

# Data
A module should work with files on the local file system with as little knowledge of the surrounding system as possible.
However, data is expected to be written/read using a particular directory structure:

## `in/data`
Any input files that your module needs will be available in the input blob directory `in/data`

## `out/data`
Any output files you wish to persist should be written to `out/data`

## `out/meta.json`
Any metadata you wish to persist should be written to the JSON file `out/meta.json`
**Metadata Schema:**
Key value pairs can only currently be stored as strings.
```json
{
  "key": "value", [optional]
  ...
  "files": "myblob.png,myblob2.png" [required]
}
```

## `out/events`
Any events you wish to publish should be stored as JSON files in `out/events`
**Event Schema:**
Key value pairs can only currently be stored as strings.
```json
{
  "key": "value", [optional]
  ...
  "eventType": "myEventType" [required]
}
```

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

## `GET /ready`
Synchronous call to setup the execution environment

**headers:**

`{"secret": "<shared_secret>"}`

