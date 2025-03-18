# see if latest-dev used when no tag is specified
FROM python

RUN apt-get update \
    && apt-get install --assume-yes --no-install-recommends \
        gettext \
        git \
        libpq5 \
        make \
        rsync
