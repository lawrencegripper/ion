import os
import time
import shutil
import datetime
import sys
import os
from lib import sidecar

# Environment variables:
# - SIDECAR_BASE_DIR [OPTIONAL]

base_dir = "/ion/"
if "SIDECAR_BASE_DIR" in os.environ:
    base_dir = os.environ["SIDECAR_BASE_DIR"]
else:
    print("SIDECAR_BASE_DIR not set, defaulting to /ion/")

print("module starting")

# This module has no input files as it
# is the first in it's graph. However,
# if it was a subsequent module it could
# read files from `{base_dir}/in/data`
# and structured data from the JSON file
# `{base_dir}/in/meta.json`

# Do some processing, this could be anything
print("fake doing some work...")
spinner = sidecar.spinning_cursor()
for _ in range(50):
    sys.stdout.write(next(spinner))
    sys.stdout.flush()
    time.sleep(0.1)
    sys.stdout.write('\b')

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

# Now we're finished
print("module finished")