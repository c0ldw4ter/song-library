FROM golang:1.23-alpine

# Apply work dir(Устанавливаем рабочую директорию)
WORKDIR /app

# Install all deps to Alpine Linux(Устанавливаем зависимости Alpine Linux)
RUN apk add --no-cache postgresql-client

# Copy code and install all deps for golang(Копируем зависимости и код)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run entrypoint.sh to migrate(Создаем entrypoint.sh для выполнения миграций)
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Built app(Собираем приложение)
RUN go build -o main .

# Command to run app(Команда для запуска приложения)
ENTRYPOINT ["/entrypoint.sh"]