FROM node:24-bookworm-slim AS build
WORKDIR /src
RUN npm install --global pnpm@10.33.0 --force
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile
COPY . ./
ARG VITE_DEDHED_AUDIENCE=private
ENV VITE_DEDHED_AUDIENCE=$VITE_DEDHED_AUDIENCE
RUN pnpm build
FROM nginx:1.31.7-alpine@sha256:892bb2152464f0c5d8972593a5f5afaf8054552580b699d95f8ff4af08c2d33d
COPY --from=build /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
