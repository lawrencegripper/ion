import requests
import os
import time
import shutil
import uuid
import datetime
import sys
from random import randint
from shutil import copyfile

"""
get_parent_meta
---
Performs a HTTP GET request against the sidecar API: /parent/meta
"""
def get_parent_meta():
    print("get_parent_meta()")
    headers = {"secret": shared_secret}
    res = requests.get(parent_meta_url, headers=headers)
    if res.status_code == 200:
        return res.json()
    print("Error requesting parent metadata: code {}".format(res.status_code))
    sys.exit()

"""
download_image_from_blob
---
Performs a HTTP GET request against the sidecar API: /parent/blob
"""
def download_image_from_blob(image_name):
    print("download_image_from_blob({})".format(image_name))
    headers = {"secret": shared_secret}
    res = requests.get("{}?res={}".format(parent_blob_url, image_name),
     headers=headers,
     stream=True)
    if res.status_code > 399:
        print("Failed to get image '{}'".format(image_name))
        print("Status code: {}, Content: {}".format(res.status_code, res.content))
        sys.exit()
    if not os.path.exists(image_dir):
        os.makedirs(image_dir)
    filename = image_name.split("/")[-1]
    image_file_path = os.path.join(image_dir, filename)
    try:
        with open(image_file_path, 'wb') as out_file:
            res.raw.decode_content = True
            shutil.copyfileobj(res.raw, out_file)
        del res
        return image_file_path
    except Exception as ex:
        print("Exception '{}' thrown whilst writing to file '{}'".format(ex, image_file_path))
        sys.exit()

"""
process_image
---
Simulates the processing of an image
"""
def process_image(image_file_path):
    print("process_image({})".format(image_file_path))
    time.sleep(10)
    return image_file_path

"""
upload_image_to_blob
---
Performs a HTTP PUT request against the sidecar API: /self/blob
"""
def upload_image_to_blob(image_file_path):
    print("upload_image_to_blob({})".format(image_file_path))
    with open(image_file_path, 'rb') as in_file:
        data = in_file.read()
        headers = {"secret": shared_secret}
        filename = image_file_path.split("/")[-1]
        rint = randint(0, 9)
        filename = str(rint) + filename
        print("uploading {}".format(filename))
        res = requests.put("{}?res={}".format(self_blob_url, filename),
            data=data,
            headers=headers,
            params={'file':filename})
        return res.text

"""
add_metadata
---
Performs a HTTP PUT request against the sidecar API: /self/meta
"""
def add_metadata(metadata):
    print("add_metadata({})".format(metadata))
    headers = {'secret': shared_secret, 'Content-Type': 'application/json'}
    res = requests.put(self_meta_url, headers=headers, data=metadata)

"""
publish_event
---
Performs a HTTP POST request against the sidecar API: /events
"""
def publish_event(event):
    print("publish_event({})".format(event))
    headers = {'secret': shared_secret, 'Content-Type': 'application/json'}
    res = requests.post(events_url, headers=headers, data=event)

"""
print_env
---
Prints the current environment
"""
def print_env():
    print()
    print("Starting Python example")
    print("...............")
    print("Shared secret: {}".format(shared_secret))
    print("Server port: {}".format(port))
    print("Sidecar endpoint: {}".format(sidecar_endpoint))
    print("Parent Blob URL: {}".format(parent_blob_url))
    print("Parent Meta URL: {}".format(parent_meta_url))
    print("My Blob URL: {}".format(self_blob_url))
    print("My Meta URL: {}".format(self_meta_url))
    print("Events URL: {}".format(events_url))
    print("...............")
    print()

# Setup
# ---
# Please ensure the sidecar is running and that it has been configured properly.
# This sample will expect the parent to have metadata set with a key 'imageName'
# with an associated value that is a legitimate blob file.
#
# Running
# ---
# SHARED_SECRET=secret SIDECAR_PORT=8080 python3 example.py
#

shared_secret = os.environ["SHARED_SECRET"]
port = os.environ["SIDECAR_PORT"]

# Global vars
sidecar_endpoint = "http://localhost:" + port
image_dir = "images"

# The parent is the previous module who fired the event
# that triggered this current module. Any data that module
# stored will be available in either it's meta store or blob store.
parent_meta_url = sidecar_endpoint + "/parent/meta"
parent_blob_url = sidecar_endpoint + "/parent/blob"

# Self is this modules storage options. We can store metadata in our
# metastore and blob data in our blob store.
self_blob_url = sidecar_endpoint + "/self/blob"
self_blobs_url = sidecar_endpoint + "/self/blobs"
self_meta_url = sidecar_endpoint + "/self/meta"

# Once we have completed our processing, we may wish to fire more
# events to trigger downstream jobs. We can use the events endpoint
# for this.
events_url = sidecar_endpoint + "/events"

# Print the current environment
print_env()

# Get parent metadata as a dictionary of string:string
parent_meta = get_parent_meta()
image_name = parent_meta['imageName']

# Download the desired image from blob storage
image_path = download_image_from_blob(image_name)

# Process the image
processed_image_path = process_image(image_path)

# Upload the image to blob storage
blob_uri = upload_image_to_blob(processed_image_path)

# Store insights as metadata
metadata = {
    "processedImageURL": blob_uri
}
add_metadata(metadata)

# Fire an event
event = {
    "timestamp": datetime.datetime.now().isoformat()
}
publish_event(event)
