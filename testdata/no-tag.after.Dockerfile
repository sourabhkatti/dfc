# see if latest-dev used when no tag is specified
FROM cgr.dev/ORG/python:latest-dev
USER root

RUN apk add -U gettext git libpq make rsync
