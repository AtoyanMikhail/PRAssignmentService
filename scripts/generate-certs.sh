#!/bin/bash

# Скрипт для генерации self-signed TLS сертификатов для разработки

set -e

CERTS_DIR="certs"
KEY_FILE="$CERTS_DIR/server.key"
CERT_FILE="$CERTS_DIR/server.crt"

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Генерация TLS сертификатов для разработки...${NC}"

# Создание директории если её нет
mkdir -p "$CERTS_DIR"

# Проверка существования сертификатов
if [ -f "$KEY_FILE" ] && [ -f "$CERT_FILE" ]; then
    echo -e "${YELLOW}Сертификаты уже существуют!${NC}"
    read -p "Перегенерировать? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Генерация отменена."
        exit 0
    fi
    echo "Удаление старых сертификатов..."
    rm -f "$KEY_FILE" "$CERT_FILE"
fi

# Генерация приватного ключа (2048 бит RSA)
echo "Генерация приватного ключа..."
openssl genrsa -out "$KEY_FILE" 2048

# Установка прав доступа на ключ (только владелец может читать)
chmod 600 "$KEY_FILE"

# Генерация самоподписанного сертификата (действителен 365 дней)
echo "Генерация самоподписанного сертификата..."
openssl req -new -x509 -sha256 \
    -key "$KEY_FILE" \
    -out "$CERT_FILE" \
    -days 365 \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=PR Assignment Service/OU=Development/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:127.0.0.1,IP:0.0.0.0"

# Установка прав доступа на сертификат
chmod 644 "$CERT_FILE"

echo -e "${GREEN}✓ Сертификаты успешно созданы!${NC}"
echo ""
echo "Файлы:"
echo "  Приватный ключ: $KEY_FILE"
echo "  Сертификат:     $CERT_FILE"
echo ""
echo "Сертификат действителен 365 дней."
echo ""
echo -e "${YELLOW}ВНИМАНИЕ: Это self-signed сертификат для разработки!${NC}"
echo "Для production используйте сертификаты от доверенного CA."
