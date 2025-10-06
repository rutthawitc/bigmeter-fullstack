# Build stage
FROM node:20-alpine AS builder

# Enable corepack for pnpm
RUN corepack enable

WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source files
COPY . .

# Build the application
# Set VITE_API_BASE_URL to empty string for relative URLs
ARG VITE_API_BASE_URL=""
ARG VITE_LOGIN_API="https://intranet.pwa.co.th/login/webservice_login6.php"
ARG VITE_ADMIN_USERNAMES="admin"

RUN pnpm build

# Production stage
FROM nginx:alpine

# Copy built files to nginx
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
