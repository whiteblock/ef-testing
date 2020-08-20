# gcloud compute scp root@logs:/var local-directory --zone instancezone 
gcloud compute scp root@logs:/var/log/syslog-ng/prelim-8a339c.log.tar.gz $PWD/ --zone "us-central1-a" --project "infra-dev-249211"
# gcloud compute scp root@ava-test-master:/root/test-stats/final.tar.gz $PWD/ --zone "us-central1-a" --project "whiteblock-infra"