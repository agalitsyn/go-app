# Пример приложения

Приложение готово к деплою в Deis.

## TL;DR

```sh
$ make docker-build

$ make docker-push

$ deis login <URL>

$ deis create <имя проекта>-<имя приложения>

$ deis config:set HEALTHCHECK_URL=/healthz
$ deis config:set HEALTHCHECK_INITIAL_DELAY=3
$ deis config:set HEALTHCHECK_TIMEOUT=10
$ deis config:set DATABASE_URL=postgres://<user>:<password>@<url>:<port>/<dbname>

$ make docker-build

# Нужно уточнить в какой registry вы будете пушить образ
$ IMAGE=my-app:latest REGISTRY=my-registry.private make docker-push-private

$ deis pull <image>
Creating build... o..

$ deis info
```

## Работа локально

В `Makefile` есть набор команд, которые облегчат разработку локально.
Для старта приложений используется [goreman](https://github.com/mattn/goreman).

Запуск приложения локально:

```sh
$ make run
```

*Note:* в приложении используется база данный postgresql. Доступ к ней прописывается через ENV, в файле `.env`.

Запустить postgres для разработки можно в докере
```
docker run --name postgres94 -e POSTGRES_PASSWORD=docker -e POSTGRES_USER=docker -p 5432:5432 postgres:9.4
```

## Деплой в Deis

[Начало работы](http://docs.deis.io/en/latest/using_deis/install-client/).

Деплой в Deis сделан через [Docker image](http://docs.deis.io/en/latest/using_deis/using-docker-images/#using-docker-images).

Плюсы такого решения:
* Полный контроль над сборкой приложения
* Полный контроль над запуском приложения
* Можно дистрибьютить, артефакт сборки доступен любому члену команды.
* Можно делать CI цепочку с тестированием одно и того же артефакта, именно этот артефакт в итоге будет запущен на продакшене.

Работа с образами оптимизированна, используется пустой base image - scratch.
[Дополнительно](https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07#.rhfm1i8ug).


### Сборка приложения

Нужно собрать статический бинарник.

Обычно, если выполнить команду `ldd` на артефакт сборки, то видно использование динамической линковки:
```
linux-vdso.so.1 => (0x00007fff039fe000)
libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f61df30f000)
libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f61def84000)
/lib64/ld-linux-x86-64.so.2 (0x00007f61df530000)
```

Очевидно, это не будет работать в `scratch` контейнере, так как там нет операционной системы с этими бибилиотеками.
Команда `make build` выставит нужные флаги для линкера. Подробнее см [тикет](https://github.com/golang/go/issues/9344#issuecomment-69944514).

### Деплой контейнера

Команда `make docker-build` соберет бинарник и подложит его в контейнер.

Далее нужно запушить контейнер, делаем `make docker-push`, либо пушим в приватный registry не забыв указать ENV
переменные:
```
IMAGE=my-app:latest REGISTRY=my-registry.private make docker-push-private
```

### Healthcheck

Конфигурируем healthcheck

```sh
$ deis config:set HEALTHCHECK_URL=/healthz
$ deis config:set HEALTHCHECK_INITIAL_DELAY=3
$ deis config:set HEALTHCHECK_TIMEOUT=10
```

Подробнее в [документации](http://docs.deis.io/en/latest/using_deis/config-application/#custom-health-checks).

