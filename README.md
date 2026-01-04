# Metrics Collection Service

Сервис сбора и хранения метрик.

## Компоненты

- **Server** - HTTP-сервер для приёма и хранения метрик
- **Agent** - агент для сбора системных метрик и отправки на сервер

## Запуск

```bash
make up
```

## Документация

- [Swagger API документация](http://localhost:8080/swagger/) - доступна после запуска сервера
- [Документация пакетов](http://localhost:6060/pkg/github.com/arvaliullin/metrics-collection-service/?m=all) - запуск: `make godoc`
