# piggytalk-backend

这是一个readme

## docker cmd

```docker network create -d bridge piggytalk-backend-bridge```
| container | command                                                                                                                                                                     |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| consul    | ```docker run -d -p 8500:8500 --net=piggytalk-backend-bridge -e CONSUL_BIND_INTERFACE='eth0' --name=consul consul agent -server -bootstrap -ui -node=1 -client='0.0.0.0'``` |


## service info

*所有服务的http接口应为8xxx，grpc接口应为9xxx*

| service   | http port | grpc port | service name        |
|-----------|-----------|-----------|---------------------|
| snowflake | 8000      | 9000      | piggytalk-snowflake |
| gateway   | 8080      | *none*    | *none*              |
