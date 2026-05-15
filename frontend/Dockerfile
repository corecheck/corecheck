FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json .
COPY . .
RUN npm install
RUN npm run build
RUN npm prune --production

FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/build build/
COPY --from=builder /app/node_modules node_modules/
COPY package.json .
EXPOSE 3000
ENV NODE_ENV=production \
    PUBLIC_DASHBOARD_GITHUB_URL="" \
    PUBLIC_DASHBOARD_TESTS_URL="" \
    PUBLIC_DASHBOARD_BENCHMARKS_URL="" \
    PUBLIC_DASHBOARD_JOBS_URL=""
CMD [ "node", "build" ]
