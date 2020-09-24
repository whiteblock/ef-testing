"""download_parsed.py

Downloads parsed logs (block propagation and reorg log lines only) from Google
Cloud Storage.
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

    for folder in folders:
        if folder in IGNORE_FOLDERS:
            # skip
            continue

        blob = bucket.blob(f"{folder}/autoexec.log")
        blob.download_to_filename(f"autoexec.log")

        with open("autoexec.log", 'r') as f:
            for line in f:
                # new process
                test = json.loads(line)
                name = test["testName"]
                test_id = test["testID"]
                os.system(f"gsutil -m cp -R gs://whiteblock-logs/{folder}/{name}_{test_id} .")