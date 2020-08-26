# Set TEST_ID here
TEST_ID=

if [[ $1 -ne 0 ]]; then
    TEST_ID=$1
    echo hi
fi

if [[ -z ${TEST_ID} ]]; then
    echo "need to set TEST_ID variable in shell script are as bash argument"
    exit
fi

echo "Downloading syslog-ng logs and resource stats for ${TEST_ID}"
gcloud compute scp --compress --recurse root@logs:/var/log/syslog-ng/backups/${TEST_ID} $PWD/ --zone "us-central1-a" --project "infra-dev-249211"
gcloud compute scp --compress --recurse root@logs:/var/log/syslog-ng/backups/${TEST_ID}.stats $PWD/ --zone "us-central1-a" --project "infra-dev-249211"
# gcloud compute scp --compress root@bill-dev-vm-1:/test/ef-testing/stats/rstats-abd345c0-79f2-4b96-91ea-75d924dcfec8 $PWD/ --zone "us-central1-a" --project "infra-dev-249211"