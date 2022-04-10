# message-queue

Для запуска серверной части и обработчика:
`docker-compose up  --scale worker=4 rabbitmq worker pathfinder`

После `worker` указать число желаемых работников.

Для запуска клиента: 
`go run ./cmd/client/main.go`