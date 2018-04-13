import requests
import os
import time
import json
import sys
import shutil

def ready(port, shared_secret):
    headers = {"secret": shared_secret}
    count = 0
    retryCount = 5
    retryDelay = 5
    for i in range(0, retryCount):
        try:
            res = requests.get("http://localhost:{}/ready".format(port), headers=headers)
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

def done(port, shared_secret):
    headers = {"secret": shared_secret}
    res = requests.get("http://localhost:{}/done".format(port), headers=headers)
    if res.status_code != 200:
        print("Failed to commit state")
        body = json.loads(res.text)
        print("response: {}".format(body))
        sys.exit(1)

def write_file(base_dir, filename, text):
    path = "{}/out/data/{}".format(base_dir, filename)
    with open(path, "w") as out:
        out.write(text)

def write_event(base_dir, filename, event):
    path = "{}/out/events/{}".format(base_dir, filename)
    with open(path, 'w') as out:
        json.dump(event, out)

def write_insight(base_dir, insight):
    path = "{}/out/meta.json".format(base_dir)
    with open(path, 'w') as out:
        json.dump(insight, out)

def refresh(base_dir):
    if os.path.exists(base_dir):
        shutil.rmtree(base_dir, ignore_errors=True)
    os.makedirs(base_dir)
    os.makedirs("{}/in/data".format(base_dir))
    os.makedirs("{}/out/data".format(base_dir))
    os.makedirs("{}/out/events".format(base_dir))