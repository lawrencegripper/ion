import requests
import os
import time
import shutil
import uuid
import datetime
import sys
from shutil import copyfile

shared_secret = os.environ["SHARED_SECRET"]
port = os.environ["SIDECAR_PORT"]

sidecar_endpoint = "http://localhost:" + port
blob_inputs_url = sidecar_endpoint + "/blob/inputs"
blob_outputs_url = sidecar_endpoint + "/blob/outputs"
meta_inputs_url = sidecar_endpoint + "/meta/inputs"
meta_outputs_url = sidecar_endpoint + "/meta/outputs"
events_url = sidecar_endpoint + "/events"

print()
print("Starting sample")
print("...............")
print("Shared secret: {}".format(shared_secret))
print("Server port: {}".format(port))
print("Sidecar endpoint: {}".format(sidecar_endpoint))
print("Blob inputs URL: {}".format(blob_inputs_url))
print("Blob outputs URL: {}".format(blob_outputs_url))
print("Metadata inputs URL: {}".format(meta_inputs_url))
print("Metadata outputs URL: {}".format(meta_outputs_url))
print("Events URL: {}".format(events_url))
print("...............")
print()

# Get input metadata
print("Getting input metdata...")
headers = {'secret': shared_secret}
res = requests.get(meta_inputs_url, headers=headers)
#TODO: error check
input_metadata = res.json()
print("Metadata: {}".format(input_metadata))
image_url = input_metadata['imageURL']
print("Image URL: {}".format(image_url))
del res

# Get authenticated URL to image
print("Getting SAS URL to download image...")
res = requests.get("{}?url={}".format(blob_inputs_url, image_url), headers=headers)
#TODO: error check
image_sas_url = res.text
print("SAS URL: {}".format(image_sas_url))
del res

# Download image from blob
print("Downloading image from Blob and writing it to disk...")
res = requests.get(image_sas_url, stream=True)
if res.status_code > 400:
    print("Failed to get image from url '{}'".format(image_sas_url))
    print("Status code: {}, Content: {}".format(res.status_code, res.content))
    sys.exit()
image_dir = "images"
if not os.path.exists(image_dir):
    os.makedirs(image_dir)
filename = image_url.split("/")[-1]
image_file_path = os.path.join(image_dir, filename)
try:
    with open(image_file_path, 'wb') as out_file:
        res.raw.decode_content = True
        shutil.copyfileobj(res.raw, out_file)
    del res
except Exception as ex:
    print("Exception '{}' thrown whilst writing to file '{}'".format(ex, image_file_path))
    sys.exit()

# Process images
print("Processing image...")
time.sleep(10)

# Get new blob URL
print("Get an output blob URL...")
res = requests.get(blob_outputs_url, headers=headers)
#TODO: error check
blob_output_auth_url = res.text
print("Output blob URL: {}".format(blob_output_auth_url))
del res
del headers

segs = blob_output_auth_url.split('?', 1)
blob_base_url = segs[0]
blob_url = blob_base_url + "/" + image_file_path
blob_auth_url = blob_url + "?" + segs[1]

# Store output in blob
print("Uploading processed image to Blob...")
print("Blob upload URL: {}".format(blob_auth_url))
with open(image_file_path, 'rb') as in_file:
    data = in_file.read()
    headers = {}
    res = requests.put(blob_auth_url,
                    data=data,
                    headers={'x-ms-blob-type':'BlockBlob'},
                    params={'file':image_file_path}
                )

# Store metadata
print("Storing metadata about processing...")
del headers
data = {
            "objectDetectedCount": "200",
            "imageURL": blob_url
        }
headers = {'secret': shared_secret, 'Content-Type': 'application/json'}
res = requests.post(meta_outputs_url, headers=headers, data=data)
#TODO: error check
del res
del headers

# Publish events
data = {
            "some_key": "some_data",
            "timestamp": datetime.datetime.now().isoformat(),
        }
headers = {'secret': shared_secret, 'Content-Type': 'application/json'}
res = requests.post(events_url, headers=headers, data=data)
#TODO: error check