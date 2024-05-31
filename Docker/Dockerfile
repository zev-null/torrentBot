# Первый этап: сборка приложения
FROM ubuntu:22.04 as builder

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Установка Go
RUN curl -fsSL https://go.dev/dl/go1.21.6.linux-amd64.tar.gz -o go.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

# Копирование исходных файлов проекта и сборка приложения
COPY . .
RUN /usr/local/go/bin/go mod download
RUN /usr/local/go/bin/go build -o target/bot main.go 

# Второй этап: создание минимального образа для запуска приложения
FROM ubuntu:22.04
RUN mkdir /opt/bot \
    && useradd -u 4242 --shell /bin/bash --home /opt/bot bot \
    && chmod u+rwx /opt/bot

RUN chown -R bot /opt/bot
RUN apt-get update && apt-get install -y ca-certificates

USER bot
# Копирование только бинарного файла из первого этапа
COPY --from=builder /app/target/bot /opt/bot/

CMD ["/opt/bot/bot"]
