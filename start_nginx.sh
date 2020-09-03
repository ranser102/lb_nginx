#/bin/bash
nginx -g "daemon off;" && /etc/nginx/HEALTHCHECK/health-check /etc/nginx/HEALTHCHECK/health-check.json /etc/nginx/HEALTHCHECK/health-check.log
