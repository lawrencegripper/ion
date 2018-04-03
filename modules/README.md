# Module
A module is a processing unit or task. It can be written in any language as long as it has all its dependencies packaged into a container. All modules must be hosted on an accessible container registry in order for the Dispatcher to schedule it.

## Envrionment variables
Each module will be supplied with the following environment varaibles:
* `SHARED_SECRET` - Used to authenticate requests to the sidecar
* `SIDECAR_PORT` - The local port your module can communicate with the sidecar on

## Function
The function of a module is to execute some processing on some data.
Existing data can be accessed via the sidecar API or pulled in from other sources. Please refer to the [Sidecar docs](../sidecar/README.md) for further details.
Once you have the data you need, you are free to process it however you see fit.
Once complete, you'll probably want to store the output in some external storage too.
Again, this is possible using the sidecar or an directly against an external endpoint i.e. Azure SQL DB.

> **Please note:** data you want to be accessible between independent modules should be stored or referenced using the provided sidecar data stores.

