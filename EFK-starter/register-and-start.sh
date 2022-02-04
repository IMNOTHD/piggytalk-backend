#!/bin/bash

curl --location -g --request PUT 'http://consul:8500/v1/agent/service/register'  \
        --data-raw '{"id": "fluentd1","name": "fluentd","port": 24224,"check": {"name": "Fluentd TCP on port 5672","tcp": "fluentd:24224","interval": "10s","timeout": "1s","DeregisterCriticalServiceAfter": "10s"}}'

fluentd