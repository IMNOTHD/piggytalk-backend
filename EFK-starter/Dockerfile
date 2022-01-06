FROM fluent/fluentd:v1.14-debian-1
# TODO change docker image to arm version before upload to oracle could

# Use root account to use apt
USER root

RUN echo \
    deb http://mirrors.aliyun.com/debian/ buster main non-free contrib\
    deb-src http://mirrors.aliyun.com/debian/ buster main non-free contrib\
    deb http://mirrors.aliyun.com/debian-security buster/updates main\
    deb-src http://mirrors.aliyun.com/debian-security buster/updates main\
    deb http://mirrors.aliyun.com/debian/ buster-updates main non-free contrib\
    deb-src http://mirrors.aliyun.com/debian/ buster-updates main non-free contrib\
    deb http://mirrors.aliyun.com/debian/ buster-backports main non-free contrib\
    deb-src http://mirrors.aliyun.com/debian/ buster-backports main non-free contrib\
    > /etc/apt/sources.list

# below RUN includes plugin as examples elasticsearch is not required
# you may customize including plugins as you wish
RUN buildDeps="sudo make gcc g++ libc-dev" \
 && apt-get update \
 && apt-get install -y --no-install-recommends $buildDeps \
 && sudo gem install fluent-plugin-elasticsearch \
 && sudo gem sources --clear-all \
 && SUDO_FORCE_REMOVE=yes \
    apt-get purge -y --auto-remove \
                  -o APT::AutoRemove::RecommendsImportant=false \
                  $buildDeps \
 && rm -rf /var/lib/apt/lists/* \
 && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem

RUN mkdir -p /fluentd/etc
COPY fluent.conf /fluentd/etc/
COPY fluent.conf /etc/fluent/fluent.conf
COPY entrypoint.sh /bin/

ENV FLUENTD_CONF="fluent.conf"

RUN apt-get update; apt-get install -qq curl -y
COPY register-and-start.sh /register-and-start.sh
RUN chmod +x register-and-start.sh

USER fluent

CMD ["/register-and-start.sh"]
