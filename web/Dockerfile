FROM node:24-bookworm-slim AS build
WORKDIR /src
RUN npm install --global pnpm@10.33.0 --force
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile
COPY . ./
RUN pnpm build
FROM nginx:1.29-alpine
COPY --from=build /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
