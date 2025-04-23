FROM cgr.dev/ORG/python:3.9-dev
USER root

RUN echo "STEP 1" && \
    apk add --no-cache py3-pip py3-virtualenv python-3 && \
    echo "STEP 2" && \
    echo "STEP 3" && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* ~/.cache ~/.npm

RUN echo hello

RUN echo hello && \
    echo goodbye

RUN apk add --no-cache py3-pip py3-virtualenv python-3

RUN true
