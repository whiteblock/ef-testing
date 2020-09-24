""" concat_autoexeclog.py

Concatenates autoexec.log files across multiple directories in Google Cloud
Storage.
"""
import os
import json
from google.cloud import storage

client = storage.Client()
bucket = client.get_bucket("whiteblock-logs")
IGNORE_FOLDERS = ["ef-logs-prelim"]


if __name__ == '__main__':
    blobs = bucket.list_blobs()
    paths = [blob.name for blob in blobs]

    # get all folders
    folders = []
    for path in paths:
        if path.split('/')[0] not in folders:
            folders.append(path.split('/')[0])

    test_info = []
    for folder in folders:
        if folder in IGNORE_FOLDERS:
            # skip
            continue

        blob = bucket.blob(f"{folder}/autoexec.log")
        blob.download_to_filename(f"autoexec.log")

        with open("autoexec.log", 'r') as f:
            for line in f:
                test_info.append(json.loads(line))


    test_info = sorted(test_info, key = lambda x: x["testName"])
    with open("master_autoexec.log", "w") as master:
        for item in test_info:
            master.write(json.dumps(item) + "\n")
