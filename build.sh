#!/bin/bash

# Функция для отображения помощи
show_help() {
    echo "Использование: ./build.sh [опция]"
    echo ""
    echo "Опции:"
    echo "  local     - Собрать только Go приложение локально"
    echo "  docker    - Собрать Docker образ с локальным бинарником"
    echo "  full      - Полная сборка Docker образа с компиляцией внутри"
    echo "  help      - Показать эту справку"
    echo ""
}

build_local() {
    echo "🔨 Собираем Go приложение локально..."
    cd app
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s' -o main ./cmd/main.go

    if [ $? -eq 0 ]; then
        echo "✅ Приложение успешно собрано!"
        echo "📁 Бинарный файл: app/main"
    else
        echo "❌ Ошибка сборки приложения"
        exit 1
    fi
}

build_docker() {
    echo "🐳 Собираем Docker образ с готовым бинарником..."
    build_local
    docker build -t googlesheetw:latest .

    if [ $? -eq 0 ]; then
        echo "✅ Docker образ успешно собран!"
        echo "🚀 Запуск: docker-compose up"
    else
        echo "❌ Ошибка сборки Docker образа"
        exit 1
    fi
}

build_full() {
    echo " Полная сборка Docker образа..."
    docker build -f dockerfile-prebuild -t googlesheetw:latest .

    if [ $? -eq 0 ]; then
        echo "✅ Docker образ успешно собран!"
        echo "🚀 Запуск: docker-compose up"
    else
        echo "❌ Ошибка полной сборки Docker образа"
        exit 1
    fi
}

# Проверяем аргументы
case "${1:-docker}" in
    local)
        build_local
        ;;
    docker)
        build_docker
        ;;
    full)
        build_full
        ;;
    help)
        show_help
        ;;
    *)
        echo "❌ Неизвестная опция: $1"
        show_help
        exit 1
        ;;
esac
