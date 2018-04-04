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
    print("Commit URL: {}".format(commit_url))
    print("...............")
    print()

def ready():
    headers = {"secret": shared_secret}
    count = 0
    retryCount = 5
    retryDelay = 5
    for i in range(0, retryCount):
        try:
            res = requests.get(ready_url, headers=headers)
            if res.status_code != 200:
                print("Sidecar could not become ready")
                sys.exit(1)
            return
        except Exception:
            count += 1
            print("Sidecar connect retry count: {}".format(count))
            if count < retryCount:
                time.sleep(retryDelay)
    print("failed to connect to sidecar, is it running!?")
    sys.exit(1)


def commit():
    headers = {"secret", shared_secret}
    res = requests.get(commit_url, headers=headers)
    if res.status_code != 200:
        print("Failed to commit state")
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

# Global vars
sidecar_endpoint = "http://localhost:" + port
image_dir = "images"
input_dir = "in/data/"
output_dir = "out/data/"
events_dir = "out/events/"
in_meta_path = "in/meta.json"
out_meta_path = "out/meta.json"
ready_url = sidecar_endpoint + "/ready"
commit_url = sidecar_endpoint + "/commit"

# Test whether the sidecar is ready
ready()

# Print the current environment
print_env()

# Get input image
data = ""
in_file = "imagein.png"
with open(os.path.join(input_dir, in_file), 'r') as inf:
    data = inf.read()

# Process the image
time.sleep(10)

# Write output files
for i in range(0, 5):
    out_file = "imageout" + str(i) + ".png"
    with open(os.path.join(output_dir, out_file), "rw") as outf:
        outf.write("face!")
    # Fire an event per face
    event = {
        "eventType": "face_detected",
        "files": out_file
    }
    with open(os.path.join(events_dir, "event", str(i)), 'w') as evf:
        json.dump(data, evf)

# Write metadata
meta = {
    "source": "Facebook"
}
with open(out_meta_path, 'w') as mf:
    json.dump(data, mf)

# Commit blobs, meta and events
commit()