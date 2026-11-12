FROM php:8.3-cli-alpine
RUN apk add --no-cache icu-dev postgresql-dev oniguruma-dev \
    && docker-php-ext-install pdo_pgsql intl mbstring
WORKDIR /workspace
