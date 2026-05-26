FROM nginx:1.29.8-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=20 CMD ["wget", "-q", "--spider", "http://127.0.0.1:8080/livez"]
