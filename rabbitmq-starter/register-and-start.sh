#!/bin/bash

curl --location -g --request PUT 'http://consul:8500/v1/agent/service/register' \
        --data-raw '{"id": "rabbitmq1","name": "rabbitmq","port": 5672,"check": {"name": "RabbitMQ TCP on port 5672","tcp": "rabbitmq:5672","interval": "10s","timeout": "1s"}}'

rabbitmq-server