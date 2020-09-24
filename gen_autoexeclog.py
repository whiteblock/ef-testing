"""gen_autoexeclog.py

Script to generate the autoexec.log for all raw logs in a single folder on
Google Cloud Storage.
"""
import os
from os import path
import yaml
from google.cloud import storage
import json

# folder in gcp bucket
folder = ""

if __name__ == '__main__':
    client = storage.Client()
    bucket = client.get_bucket("whiteblock-logs")

    id_keys = ["_TEST0_ID", "_TEST1_ID"]
    stats_keys = ["_TEST0_STATS", "_TEST1_STATS"]
    os.chdir("test-yaml/")
    autoexec_log = []

    for file in os.listdir("."):
        with open(file) as f:
            indicator, ext = file.split(".")

            if ext != "yaml":
                continue

            test_yaml = yaml.load(f, Loader=yaml.FullLoader)
            for i in range(len(id_keys)):
                test_id = test_yaml["substitutions"][id_keys[i]]

                if test_id == "":
                    print(f"{id_keys[i]} is empty in {file}")
                    continue

                # rename syslog-ng logs
                b = bucket.get_blob(f"{folder}/{indicator}_ef-test-{test_id[:8]}.log")
                if b == None:
                    continue

                # collect all web stats from yaml files
                stats = test_yaml["substitutions"][stats_keys[i]]
                if stats == "":
                    print(f"{stats_keys[i]} is empty in {file}")
                    continue

                test_env = {
                    "hostName": "logs",
                    "testName": indicator,
                    "timeBegin": 0,
                    "timeEnd": 0,
                    "testID": test_id,
                    "webStats": json.loads(stats),
                }
                print(indicator)
                autoexec_log.append(test_env)


    with open("autoexec.log", "w") as f:
        for line in autoexec_log:
            f.write(json.dumps(line) + '\n')


    blob = bucket.blob(f"{folder}/autoexec.log")
    blob.upload_from_filename("autoexec.log")



# type TestEnv struct {
#     HostName  string     `json:"hostName"`
#     TestName  string     `json:"testName"`
#     TimeBegin int64      `json:"timeBegin"`
#     TimeEnd   int64      `json:"timeEnd"`
#     TestID    string     `json:"testID"`
#     WebStats  jsonStruct `json:"webStats"`
# }