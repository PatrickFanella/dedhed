FROM node:24-bookworm-slim AS build
WORKDIR /src
RUN npm install --global pnpm@10.33.0 --force
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile
COPY . ./
ARG VITE_DEDHED_AUDIENCE=private
ENV VITE_DEDHED_AUDIENCE=$VITE_DEDHED_AUDIENCE
RUN pnpm build
FROM nginx:1.31.8-alpine@sha256:713fc2ebfd92e8217db451477182ee91944b123d2b1ee9fd5a8c6e91ca6c3a11
COPY --from=build /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
