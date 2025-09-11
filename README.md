# capi

config api

## lint

golangci-lint run

## 安装 Postgres 和 Memcached

```
$ docker volume create postgres
$ vim docker-compose.yml
volumes:
  postgres:
    external: true

services:
  db:
    image: postgres:17.6
    restart: always
    environment:
      POSTGRES_PASSWORD: Test2u8i98ry
      POSTGRES_DB: test
    volumes:
      - postgres:/var/lib/postgresql/data:rw
    ports:
      - 5432:5432

$ docker-compose up -d
$ docker-compose logs -f

创建数据库
$ docker exec -it [ContainerID] /bin/bash
$ psql -h 127.0.0.1 -U postgres
$ CREATE DATABASE mdp_dev;
```
