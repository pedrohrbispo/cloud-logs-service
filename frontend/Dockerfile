# ───────────────────────── build stage ─────────────────────────
FROM node:22-alpine AS build
WORKDIR /app

# Install deps first (better layer caching).
COPY package*.json ./
RUN npm ci

# Build the SPA + recompute the CSP hash for the no-flash inline script.
COPY . .
RUN npm run build

# ───────────────────────── serve stage ─────────────────────────
# Unprivileged nginx: runs as UID 101, listens on 8080 (no root, no :80).
FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 8080
