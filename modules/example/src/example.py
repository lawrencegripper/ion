import requests
import os
import time
import shutil
import uuid
import datetime
import sys
import json
from random import randint
from shutil import copyfile

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
    print("Ready URL: {}".format(ready_url))
    print("Done URL: {}".format(done_url))
    print("...............")
    print()

def setup():
    if not os.path.exists(input_dir):
        os.makedirs(input_dir)

    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    if not os.path.exists(events_dir):
        os.makedirs(events_dir)

def clean_up():
    if os.path.exists(input_dir):
        shutil.rmtree(input_dir, ignore_errors=True)

    if os.path.exists(output_dir):
        shutil.rmtree(output_dir, ignore_errors=True)

    if os.path.exists(events_dir):
        shutil.rmtree(events_dir, ignore_errors=True)

def ready():
    headers = {"secret": shared_secret}
    count = 0
    retryCount = 5
    retryDelay = 5
    for i in range(0, retryCount):
        try:
            res = requests.get(ready_url, headers=headers)
            if res.status_code != 200:
                print("ready returned an error...")
                body = json.loads(res.text)
                print("response: {}".format(body))
                sys.exit(1)
            return
        except Exception:
            count += 1
            print("Sidecar connect retry count: {}".format(count))
            if count < retryCount:
                time.sleep(retryDelay)
    print("failed to connect to sidecar, is it running!?")
    sys.exit(1)


def done():
    headers = {"secret": shared_secret}
    res = requests.get(done_url, headers=headers)
    if res.status_code != 200:
        print("Failed to commit state")
        body = json.loads(res.text)
        print("response: {}".format(body))
        sys.exit(1)

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

shared_secret = ""
port = ""
base_dir = "/ion/"
if "SHARED_SECRET" in os.environ:
    shared_secret = os.environ["SHARED_SECRET"]
else:
    print("SHARED_SECRET environment variable not set!")
    sys.exit(1)

if "SIDECAR_PORT" in os.environ:
    port = os.environ["SIDECAR_PORT"]
else:
    print("SIDECAR_PORT environment variable not set!")
    sys.exit(1)

if "SIDECAR_BASE_DIR" in os.environ:
    base_dir = os.environ["SIDECAR_BASE_DIR"]
else:
    print("SIDECAR_BASE_DIR not set, defaulting to /ion/")

# Global vars
sidecar_endpoint = "http://localhost:" + port
input_dir = "{}/in/data/".format(base_dir)
output_dir = "{}/out/data/".format(base_dir)
events_dir = "{}/out/events/".format(base_dir)
in_meta_path = "{}/in/meta.json".format(base_dir)
out_meta_path = "{}/out/meta.json".format(base_dir)
ready_url = sidecar_endpoint + "/ready"
done_url = sidecar_endpoint + "/done"

clean_up()
setup()

# Test whether the sidecar is ready
ready()

# Optionally, use input data:
# unstructured files are in: in/data
# structured data is in: in/meta.json

# Print the current environment
print_env()

# Do some processing
print("doing the work...")
time.sleep(10)

# Write some output files
for i in range(0, 5):
    out_file = "image" + str(i) + ".png"
    with open(os.path.join(output_dir, out_file), "w") as outf:
        print("writing file {}".format(outf.name))
        outf.write("face!")
    # Fire an event per face
    event = [{
        "key": "eventType",
        "value": "face_detected"
    },
    {
        "key": "files",
        "value": out_file
    }]
    with open(os.path.join(events_dir, "event" + str(i) + ".json"), 'w') as evf:
        print("writing event {}".format(evf.name))
        json.dump(event, evf)

# Write insight
insight = [{
    "key": "source",
    "value": "facebook"
},
{
    "key": "imageMD5",
    "value": "1a79a4d60de6718e8e5b326e338ae533"
}]
with open(out_meta_path, 'w') as mf:
    print("writing insight {}".format(mf.name))
    json.dump(insight, mf)

# Commit blobs, meta and events
print("commiting state!")
done()
print("done")

clean_up()