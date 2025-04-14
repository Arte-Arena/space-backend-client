#!/bin/bash

echo "" >> .env
echo "PORT=$PORT" >> .env
echo "MONGODB_URI=$MONGODB_URI" >> .env
echo "TOKEN_AUDIENCE=$TOKEN_AUDIENCE" >> .env
echo "TOKEN_ISSUER=$TOKEN_ISSUER" >> .env
echo "ACCESS_TOKEN_SECRET=$ACCESS_TOKEN_SECRET" >> .env
echo "REFRESH_TOKEN_SECRET=$REFRESH_TOKEN_SECRET" >> .env
echo "ENV=$ENV" >> .env
echo "ADMIN_KEY=$ADMIN_KEY" >> .env
echo "TINY_API_TOKEN=$TINY_API_TOKEN" >> .env

echo "[arte arena security] Configurando variáveis de ambiente..."

./main
tail -f /dev/null
