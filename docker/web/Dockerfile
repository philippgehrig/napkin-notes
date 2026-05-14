# Build stage
FROM node:20-alpine AS builder

WORKDIR /build

COPY web/package.json web/yarn.lock ./
RUN yarn install --frozen-lockfile

COPY web/ .
RUN yarn build

# Runtime stage
FROM nginx:alpine

COPY --from=builder /build/dist /usr/share/nginx/html
COPY docker/web/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
