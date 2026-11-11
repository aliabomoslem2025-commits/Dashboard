FROM composer:2 AS vendor
WORKDIR /app
COPY composer.json ./
RUN composer install --no-dev --no-interaction --prefer-dist --no-scripts
FROM node:22-alpine AS assets
WORKDIR /app
COPY package.json ./
RUN npm install
COPY . .
RUN npm run build
FROM php:8.3-fpm-alpine
RUN apk add --no-cache nginx icu-dev libpq-dev oniguruma-dev && docker-php-ext-install pdo_pgsql intl mbstring
WORKDIR /var/www/html
COPY . .
COPY --from=vendor /app/vendor ./vendor
COPY --from=assets /app/public/build ./public/build
COPY docker/nginx.conf /etc/nginx/http.d/default.conf
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && mkdir -p storage/framework/cache storage/framework/sessions storage/framework/views bootstrap/cache && chown -R www-data:www-data storage bootstrap/cache
EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]
