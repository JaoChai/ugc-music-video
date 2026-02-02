# Build stage
FROM node:20-alpine AS builder
WORKDIR /app

# Accept build arguments for Vite environment variables
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL

COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Production stage
FROM node:20-alpine
WORKDIR /app
RUN npm install -g serve
COPY --from=builder /app/dist ./dist
ENV PORT=3000
CMD ["sh", "-c", "serve dist -s -l ${PORT}"]
