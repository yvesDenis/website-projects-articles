FROM node:bullseye-slim

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