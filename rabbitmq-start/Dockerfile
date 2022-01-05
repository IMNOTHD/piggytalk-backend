FROM rabbitmq:management

MAINTAINER impiggy (ideshenghe@gmail.com)


RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN sed -i s@/security.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN apt-get clean
RUN apt-get update; apt-get install curl -y
COPY register-and-start.sh /register-and-start.sh
RUN chmod +x register-and-start.sh

ENTRYPOINT ["docker-entrypoint.sh"]
EXPOSE 15691 15692 25672
CMD ["/register-and-start.sh"]
EXPOSE 15671 15672