import os
import time
import json
import sys
import shutil

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

def spinning_cursor():
    while True:
        for cursor in '|/-\\':
            yield cursor