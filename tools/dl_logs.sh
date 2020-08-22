# gcloud compute scp root@logs:/var local-directory --zone instancezone 
TEST_ID=a8f52c8e-3343-4d5b-a229-5e2f89f1d8d5

if [[ $1 -ne 0 ]]; then
    TEST_ID=$1
    echo hi
fi

if [[ -z ${TEST_ID} ]]; then
    echo "need to set TEST_ID variable"
    exit
fi

gcloud compute scp --compress --recurse root@logs:/var/log/syslog-ng/backups/${TEST_ID} $PWD/ --zone "us-central1-a" --project "infra-dev-249211"
gcloud compute scp --compress --recurse root@logs:/var/log/syslog-ng/backups/${TEST_ID}.stats $PWD/ --zone "us-central1-a" --project "infra-dev-249211"
# gcloud compute scp --compress root@bill-dev-vm-1:/test/ef-testing/stats/rstats-abd345c0-79f2-4b96-91ea-75d924dcfec8 $PWD/ --zone "us-central1-a" --project "infra-dev-249211"