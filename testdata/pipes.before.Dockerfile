FROM python:3.9.18-slim

RUN apt-get update -q -q && \
 echo "STEP 1" && \
 apt-get install python3 python3-pip python3-virtualenv --yes && \
 echo "STEP 2" && \
 apt-get -s dist-upgrade | grep "^Inst" | grep -i securi | awk -F " " {'print $2'} | xargs apt-get install --yes && \
 echo "STEP 3" && \
 apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* ~/.cache ~/.npm

RUN apt-get update -q -q && echo hello

RUN echo hello && apt-get update -q -q && echo goodbye

RUN apt-get update -q -q && \
 apt-get install python3 python3-pip python3-virtualenv --yes && \
 apt-get -s dist-upgrade | grep "^Inst" | grep -i securi | awk -F " " {'print $2'} | xargs apt-get install --yes && \
 apt-get clean

RUN apt-get update -q -q && apt-get clean
