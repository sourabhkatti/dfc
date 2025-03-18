# From https://stackoverflow.com/questions/56913746/how-to-install-python-on-nodejs-docker-image
# This Dockerfile demonstrates a Node.js app that requires Python

FROM node:9-slim

# Update apt and install Python
RUN apt-get update || : && apt-get install -y python

WORKDIR /app
COPY . /app
RUN npm install
EXPOSE 3000
CMD ["node", "index.js"] 