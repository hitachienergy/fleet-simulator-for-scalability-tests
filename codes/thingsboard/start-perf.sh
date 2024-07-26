#/bin/bash

target_QPS=80000
device_num=1111
duration=300
test_api=device

mkdir -p ~/.mytb-data && sudo chown -R 799:799 ~/.mytb-data
mkdir -p ~/.mytb-logs && sudo chown -R 799:799 ~/.mytb-logs

docker run -it --rm --network host --name tb-perf-test \
           --log-driver none \
           --env REST_URL=http://127.0.0.1:8080 \
           --env MQTT_HOST=127.0.0.1 \
           --env REST_USERNAME=tenant@thingsboard.org \
           --env REST_PASSWORD=tenant \
           --env DEVICE_END_IDX=${device_num} \
           --env MESSAGES_PER_SECOND=${target_QPS} \
           --env DURATION_IN_SECONDS=${duration} \
           --env ALARMS_PER_SECOND=1 \
           --env DEVICE_CREATE_ON_START=true \
	   --env WARMUP_ENABLED=true \
	   --env DEVICE_API=MQTT \
	   --env TEST_API=device \
           --env TEST_PAYLOAD_TYPE=SMART_METER \
	   --env TEST_ENABLED=true \
           thingsboard/tb-ce-performance-test:latest
