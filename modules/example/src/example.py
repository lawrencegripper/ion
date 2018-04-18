import os
import time
import shutil
import datetime
import sys
import os
from lib import sidecar

# NOTE:
# ---
# Please ensure the sidecar is running before you
# attempt to run this module. You must provide the
# environment variables listed below.
#
# CMD:
# ---
# SHARED_SECRET=secret SIDECAR_PORT=8080 python3 example.py
#
# ENV:
# ---
# * SHARED_SECRET [REQUIRED]
# * SIDECAR_PORT [REQUIRED]
# * SIDECAR_BASE_DIR [OPTIONAL]

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

# This line is only needed for development
# where module may be ran multiple times 
# in the same directories. This is usually 
# handled by the sidecar.
sidecar.refresh(base_dir)

print("module starting")

# Initialize the module environment.
# This must be ran before any processing
# on tracked files (files that are synced
# with Ion's data plane). This can only
# be run once.
sidecar.ready(port, shared_secret)
print("module ready")

# This module has no input files as it
# is the first in it's graph. However,
# if it was a subsequent module it could
# read files from `{base_dir}/in/data`
# and structured data from the JSON file
# `{base_dir}/in/meta.json`

# Do some processing, this could be anything
print("doing some work")
time.sleep(10)

# Write some output files
for i in range(0, 5):
    # We write the files out locally to disk.
    # Once they are sync'ed with the Ion
    # data plane, their external URIs will be
    # stored as meta data for later retrival.
    out_file = "image" + str(i) + ".png"
    sidecar.write_file(base_dir, out_file, "face!")
    print("wrote file {}".format(out_file))

    # For each file we wrote,
    # we will raise a new event
    # suggesting we found a face.
    event = [{
        "key": "eventType", # Required key
        "value": "face_detected"
    },
    {
        "key": "files",     # Required key
        "value": out_file
    }]
    event_name = "event{}.json".format(i)
    sidecar.write_event(base_dir, event_name, event)
    print("wrote event {}".format(event_name))

# During our processing, we may
# have derived some insight we 
# wish export and persist.
insight = [{
    "key": "source",
    "value": "facebook"
},
{
    "key": "image_dimensions",
    "value": "1080x1024"
},
{
    "key": "image_size",
    "value": "2.3MB"
},
{
    "key": "image_md5",
    "value": "1a79a4d60de6718e8e5b326e338ae533"
}]
sidecar.write_insight(base_dir, insight)
print("wrote new insight")

# Everything up to this point should be considered
# idempotent. If the code fails halfway through
# the module will be rescheduled and rerun.
# This call will commit our state to Ion's data plane.
# This can only be called once.
# Any state created after this call will be lost
# after the process terminates.
print("module commiting state")
sidecar.done(port, shared_secret)
print("module done")
