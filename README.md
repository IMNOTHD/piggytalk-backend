# piggytalk-backend

这是一个readme

## docker cmd

```docker network create -d bridge piggytalk-backend-bridge```

| container | command                                                                                                                                                                                                                                                                                        |
|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| consul    | ```docker run -d -p 8500:8500 --net=piggytalk-backend-bridge -e CONSUL_BIND_INTERFACE='eth0' --name=consul consul agent -server -bootstrap -ui -node=1 -client='0.0.0.0'```                                                                                                                    |
| mysql     | ```docker run --restart=always --name=mysql-piggytalk -p 3306:3306 -v :/etc/mysql/my.cnf -v :/var/lib/mysql -v :/logs -e MYSQL_ROOT_PASSWORD=123456 -d mysql:8.0 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci --secure-file-priv=/var/lib/mysql --skip-name-resolve``` |
|

## service info

*所有服务的http接口应为8xxx，grpc接口应为9xxx*

| service   | http port | grpc port | service name                |
|-----------|-----------|-----------|-----------------------------|
| snowflake | *none*    | 9000      | piggytalk-backend-snowflake |
| account   | *none*    | 9001      | piggytalk-backend-account   |
| message   | *none*    | 9002      | piggytalk-backend-message   |
| gateway   | 8080      | 9090      | *none*                      |
