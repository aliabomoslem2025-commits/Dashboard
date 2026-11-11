#!/bin/sh
set -eu
if [ -z "${APP_KEY:-}" ]; then echo 'APP_KEY is empty; set a persistent APP_KEY in the root .env file.' >&2; exit 1; fi
php artisan package:discover --ansi
php artisan migrate --force
php artisan db:seed --force
php-fpm -D
exec nginx -g 'daemon off;'
