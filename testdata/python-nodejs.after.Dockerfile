# From https://stackoverflow.com/questions/56913746/how-to-install-python-on-nodejs-docker-image
# This Dockerfile demonstrates a Node.js app that requires Python

FROM cgr.dev/ORG/node:9-dev
USER root

# Update apt and install Python
RUN : && \
    apk add --no-cache python

WORKDIR /app
COPY . /app
RUN npm install
EXPOSE 3000
CMD ["node", "index.js"] 