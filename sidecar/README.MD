# Sidecar
The sidecar is deployed alongside your module and provides a REST API for your code to interact with the bigger system. As the sidecar is ran in the same network namespace, it will be available over localhost and it's port is set via configuration. Your module can then use the sidecar's API to handle blob storage, document storage and event publishing.

# API
The REST API is based on two profiles:

* **Parent** - This is a reference to the previous module that issued the event the current module is reacting to. This allows a chain of communication to be built into your graph.

* **Self** - This is a reference to the current module.

Calls between the module and the sidecar are authenticated using a shared secret.
This is not currently used to encrypt communication between the 2 parties and is only used as a simple check.
The shared secret will be injected into your module as an environment variable: `SHARED_SECRET`.
Please add this as a header to all HTTP requests sent to the sidecar.

```python
shared_secret = os.environ["SHARED_SECRET"]
headers = {"secret": shared_secret}
res = requests.get(parent_meta_url, headers=headers)
```

## Blob Storage
Blob storage is used for handling object data, such as files.
Each module gets a dedicated directory for storing their object data, available via the sidecar's REST API.
If the blob provider you are using supports a proxy, blob requests can be proxied directly against the remote blob store, this is optimal for larger blob transfers.

## `GET /parent/blob`
Used to get blob data by name from the parent module's directory

**headers:**

`{"secret": "<shared_secret>"}`

**querystring:**

`?res=<blob_name>`

## `GET /self/blob`
Used to get blob data by name from the current module's directory

**headers:**

`{"secret": "<shared_secret>"}`

**querystring:**

`?res=<blob_name>`

## `PUT /self/blob`
Used to put blob data in the current module's directory

**headers:**

`{"secret": "<shared_secret>"}`

**querystring:**

`?res=<blob_name>`

**body:**
blob data

## `DELETE /self/blob`
Used to delete blob data from the current module's directory

**headers:**

`{"secret": "<shared_secret>"}`

**querystring:**

`?res=<blob_name>`

## `GET /self/blobs`
Used to list all the blob data in the current module's directory

**headers:**

`{"secret": "<shared_secret>"}`

**querystring:**

`?res=<blob_name>`

## Metadata Storage
Metadata storage is used for handling document data, such as maps, arrays and key-value pairs.

## `GET /parent/meta`
Used to get metadata stored by the parent module

**headers:**

{"secret": "<shared_secret>"}

## `GET /self/meta`
Used to get metadata stored by the current module

**headers:**

{"secret": "<shared_secret>"}

## `PUT /self/meta`
Used to store metadata as the current module

**headers:**

{"secret": "<shared_secret>",
"content-type": "application/json"}

**body:**
meta data

## Events
Ion is an event driven system. Modules publish events and Dispatchers consume them.
If you wish to kick off further downstream processing from your module, you can simply publish a new event to the `events` endpoint.
The event should be a dictionary/map of string values. Each event **must** have a key `eventType` and a string identifying the type of
event it is.

```python
event = {
    ...
    "eventType": "image_resize_complete"
}

```

## `POST /events`
Used to publish a new event

**headers:**

{"secret": "<shared_secret>",
"content-type": "application/json"}

**body:**
event data