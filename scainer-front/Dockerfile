# Use Node to build the SPA, then serve via nginx with /api proxy to scainer.
FROM node:22-alpine AS build
WORKDIR /app/scainer-front

COPY scainer-front/package*.json ./
RUN npm ci

COPY scainer-front/ .
COPY scainer/api/openapi.yaml ../scainer/api/openapi.yaml

RUN npm run build

FROM nginx:alpine
COPY --from=build /app/scainer-front/dist /opt/build
COPY scainer-front/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
