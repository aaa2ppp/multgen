# Multiplier Generator Service

Шаблон HTTP-сервиса и CLI-утилиты для генерации числовых множителей.

## Сборка

```bash
make all
```

Бинарники будут созданы в папке bin/.

## Запуск

**Запуск HTTP-сервера:**

```bash
bin/multgen -rtp=0.95
```

**Запуск в CLI-режиме:**

```bash
echo 10 | bin/multgen -rtp=0.95 -cli
```

**Помощь по флагам:**

```bash
bin/multgen -help
```

## Структура

- `cmd/multgen/` - основной исполняемый файл
- `internal/api/` - HTTP API обработчики
- `internal/config/` - конфигурация и парсинг флагов
- `internal/solver/` - алгоритмы генерации множителей


*Реализация алгоритма находится в стадии разработки.*