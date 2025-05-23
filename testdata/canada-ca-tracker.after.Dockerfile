# from https://github.com/canada-ca/tracker/blob/master/frontend/Dockerfile

FROM cgr.dev/ORG/node:20.16-dev AS build-env

WORKDIR /app

# Copy in whatever isn't filtered by .dockerignore
COPY . .

RUN npm ci && npm run build && npm prune --production

FROM cgr.dev/ORG/node:latest

ENV HOST 0.0.0.0
ENV PORT 3000

WORKDIR /app

COPY --from=build-env /app .

ENV NODE_ENV production
# https://github.com/webpack/webpack/issues/14532#issuecomment-947012063
ENV NODE_OPTIONS=--openssl-legacy-provider

USER nonroot
EXPOSE 3000

CMD ["index.js"]
