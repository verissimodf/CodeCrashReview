# Frontend Dockerfile - Multi-stage build

# Base stage
FROM node:20-alpine AS base

WORKDIR /app

# Install pnpm
RUN npm install -g pnpm

# Development stage
FROM base AS development

ENV NODE_ENV=development

COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY . .

EXPOSE 5173

CMD ["pnpm", "run", "dev", "--host", "0.0.0.0"]

# Build stage
FROM base AS builder

ENV NODE_ENV=production

COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile --prod=false

COPY . .

# Build production bundle
RUN pnpm run build

# Production stage (Nginx)
FROM nginx:alpine AS production

# Copy built assets
COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]