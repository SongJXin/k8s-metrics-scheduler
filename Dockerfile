FROM debian:stretch-slim

WORKDIR /

COPY  --chmod=0755 metrics-scheduler-plugin /usr/local/bin/

CMD ["metrics-scheduler-plugin"]