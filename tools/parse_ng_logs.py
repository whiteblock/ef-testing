import os
import yaml
from google.cloud import storage
import json
import queue
import threading
import time

WORKERS = 8
q = queue.Queue()
client = storage.Client()
bucket = client.get_bucket("whiteblock-logs")
PREFIX = "ef-test"

# designate folders here, e.g. ["ef-test1", "ef-auto9"]
FOLDERS = []


def do_parsing():
    print("Starting worker...")
    while True:
        test, folder = q.get()
        name = test["testName"]
        test_id = test["testID"]

        b = bucket.list_blobs(prefix=f"{folder}/{name}_{test_id}/")
        if list(b) != []:
            print(f"{name}_{test_id} already parsed, skipping...")
            q.task_done()
            continue

        syslog_ng_file = f"{name}_ef-test-{test_id[:8]}.log"
        print(f"processing {syslog_ng_file}")
        b = bucket.get_blob(f"{folder}/{syslog_ng_file}")

        if b is None:
            print(f"{syslog_ng_file} doesn't exist in gcp")
            q.task_done()
            continue

        os.system(f"time gsutil -o 'GSUtil:parallel_thread_count=1' -o 'GSUtil:sliced_object_download_max_components=8' cp gs://whiteblock-logs/{folder}/{syslog_ng_file} ./")
        os.system(f"time ./parser -t {test_id} {syslog_ng_file} {name}_{test_id}/")
        print(f"done parsing {syslog_ng_file}, uploading {name}_{test_id}/ now...")
        os.system(f"gsutil -m cp -R {name}_{test_id}/ gs://whiteblock-logs/{folder}")
        os.system(f"rm {syslog_ng_file}")
        print(f"removed {syslog_ng_file}")
        q.task_done()


if __name__ == '__main__':
    for worker in range(WORKERS):
        threading.Thread(target=do_parsing, daemon=True).start()

    blobs = bucket.list_blobs(prefix=PREFIX)
    paths = [blob.name for blob in blobs]

    for folder in FOLDERS:
        print(f"changing to folder {folder}")
        blob = bucket.blob(f"{folder}/autoexec.log")
        blob.download_to_filename(f"autoexec.log")

        with open("autoexec.log", 'r') as f:
            for line in f:
                # new process
                test = json.loads(line)
                q.put((test, folder))

            # wait for all jobs to process
            q.join()

    exit()
