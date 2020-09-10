import os
import yaml
from google.cloud import storage
import json
import queue
import threading
import time
import sys

WORKERS = 1
q = queue.Queue()
client = storage.Client()
bucket = client.get_bucket("whiteblock-logs")


def do_parsing():
    print("Starting worker...")
    while True:
        test, folder = q.get()
        name = test["testName"]
        test_id = test["testID"]

        b = bucket.get_blob(f"{folder}/{name}_{test_id}")
        if b is None:
            print(f"{test_id} already parsed, skipping...")
            q.task_done()

        syslog_ng_file = f"{name}_ef-test-{test_id[:8]}.log"
        print(f"processing {syslog_ng_file}")
        b = bucket.get_blob(f"{folder}/{syslog_ng_file}")

        if b is None:
            print(f"{syslog_ng_file} doesn't exist in gcp")
            q.task_done()

        b.download_to_filename(syslog_ng_file)
        # os.system(f"./parser -t {test_id} {syslog_ng_file} {name}_{test_id}/")
        # os.system(f"gsutil -m cp -R -Z {name}_{test_id}/ gs://whiteblock-logs/{folder}/{name}_{test_id}")
        os.system(f"rm {syslog_ng_file}")
        print(f"removed {syslog_ng_file}")
        sys.exit(1)
        q.task_done()


if __name__ == '__main__':
    for worker in range(WORKERS):
        threading.Thread(target=do_parsing).start()

    blobs = bucket.list_blobs(prefix="ef-test")
    paths = [blob.name for blob in blobs]

    # get all folders
    folders = []
    for path in paths:
        if path.split('/')[0] not in folders:
            folders.append(path.split('/')[0])

    for folder in folders:
        blob = bucket.blob(f"{folder}/autoexec.log")
        blob.download_to_filename(f"autoexec.log")

        with open("autoexec.log", 'r') as f:
            for line in f:
                # new process
                test = json.loads(line)
                q.put((test, folder))

            # wait for all jobs to process
            q.join()
