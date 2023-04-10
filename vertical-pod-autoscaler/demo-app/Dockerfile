FROM node:16@sha256:d70de5be007866473408af5cc39f6039d990d8d4f72ae13438b0b4f83b904b31

LABEL maintainer="Yves-Denis Deffo"  \
      version="1.0.0" \
      description="Ordinary http server with one exposed endpoint"

# Create app directory
WORKDIR /usr/src/app

# Install app dependencies
# A wildcard is used to ensure both package.json AND package-lock.json are copied
# where available (npm@5+)
COPY package*.json ./

RUN npm install
# If you are building your code for production
# RUN npm ci --omit=dev

# Bundle app source
COPY . .

EXPOSE 3000

ENTRYPOINT [ "node", "server.js" ]