# Генерация TLS сертификатов для разработки

Для локальной разработки используйте self-signed сертификаты.

## Быстрая генерация

Используйте готовый скрипт:

```bash
./scripts/generate-certs.sh
```

Скрипт автоматически:
- Создаст директорию `certs/` если её нет
- Сгенерирует приватный ключ (RSA 2048 бит)
- Сгенерирует самоподписанный сертификат (действителен 365 дней)
- Настроит правильные права доступа (600 для ключа, 644 для сертификата)
- Добавит Subject Alternative Names (localhost, 127.0.0.1)

## Ручная генерация сертификатов

Если предпочитаете генерировать вручную:

```bash
# Генерация приватного ключа
openssl genrsa -out certs/server.key 2048

# Генерация самоподписанного сертификата (действителен 365 дней)
openssl req -new -x509 -sha256 -key certs/server.key -out certs/server.crt -days 365 \
  -subj "/C=RU/ST=Moscow/L=Moscow/O=PR Assignment Service/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

# Установка правильных прав доступа
chmod 600 certs/server.key
chmod 644 certs/server.crt
```

## Для production

В production используйте сертификаты от доверенного CA (Let's Encrypt, AWS Certificate Manager и т.д.)

Установите переменные окружения:
```
SERVER_TLS_CERT_FILE=/path/to/production/cert.crt
SERVER_TLS_KEY_FILE=/path/to/production/cert.key
```

## Проверка сертификата

```bash
# Просмотр информации о сертификате
openssl x509 -in certs/server.crt -text -noout

# Проверка соответствия ключа и сертификата
openssl rsa -noout -modulus -in certs/server.key | openssl md5
openssl x509 -noout -modulus -in certs/server.crt | openssl md5
# Хеши должны совпадать
```

