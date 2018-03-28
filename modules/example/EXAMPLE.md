# Python Example
This Python example demonstrates how you can leverage the functionality provided by the Sidecar in order to execute processing tasks.

### Scenario
This example follows the scenario:
* Get any metadata stored by the parent module.
* Extract a image from the metadata.
* Download the image from a blob store.
* Process the image, possibly reducing it's size.
* Upload the image into a blob store.
* Store metadata about the image, i.e. new file path, new file size.
* Publish an event to trigger 1 or more subsequent jobs to do another job on the image.

### Prerequisites
* Python3
* Sidecar binary (can build from source or use container)

### Inputs
In order to run correctly, the Python module requires environment variables set.
* `SHARED_SECRET` - A pseudorandom string used to authenticate calls between the module and it's sidecar.
* `SIDECAR_PORT` - The port the sidecar is currently listening on.

### Usage
1. Ensure the Sidecar is running
2. Run `SHARED_SECRET=secret SIDECAR_PORT=8080 python3 example.py`