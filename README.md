GoProxy — это лёгкий reverse proxy и простой load balancer, написанный на Go.  
Он принимает HTTP-запросы от клиентов и распределяет их между несколькими backend-серверами
с использованием разных стратегий балансировки нагрузки.

Проект создан, чтобы продемонстрировать:

- работу с `net/http` как сервером и клиентом;
- реализацию reverse proxy;
- алгоритмы балансировки нагрузки;
- health-check backend-ов;
- конкурентность в Go (goroutines, mutexes);
- организацию проекта на Go с использованием `internal/`.

---

## Возможности

-  **Load balancing**:
    - `round_robin`
    - `least_connections`

-  **Health-check backend-серверов**
    - регулярные проверки `/health`
    - исключение "павших" серверов из списка рабочих

- **Reverse proxy**
    - пересылка любых HTTP-запросов на backend-ы
    - поддержка контекстов, таймаутов, ошибок

- **Метрики backend-ов**
    - количество запросов
    - активные соединения
    - количество ошибок
    - статус (UP/DOWN)

- **Гибкая конфигурация через `config.json`**

---

##  Структура проекта

```text
goproxy/
├── cmd/
│   └── goproxy/
│       └── main.go
│
├── internal/
│   ├── api/
│   │   └── server.go
│   │
│   ├── backend/
│   │   └── backend.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   ├── health/
│   │   └── checker.go
│   │
│   └── logger/
│       └── logger.go
│
├── example_backend/        # Мини-бэкенд для тестирования прокси
│   └── backend.go
│
├── config.json.example     # Пример конфигурации
├── go.mod
└── README.md
