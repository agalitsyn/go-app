# goapi

По мотивам [статьи](https://medium.com/@kelseyhightower/12-fractured-apps-1080c73d481c#.ihna7diaw)

В целом структура проекта имеет схему:
```
github.com/agalitsyn/foo
    foo.go      // package foo
    foo_test.go // package foo
    cmd/
        foo/
            main.go      // package main
            main_test.go // package main
```

Получается инвертированная структура, когда код библиотеки кладётся в корень,
а во вложенной папке `cmd/foo/` хранится код исполняемых программ.

Промежуточный уровень `cmd/` удобен по двум причинам:
* Инструментарий Go автоматически именует двоичные файлы по названию папки,
в которой находится package `main`, так что мы получаем наилучшие имена файлов без
возможных конфликтов с другими пакетами в репозитории.
* Если ваши пользователи применяют `go get` на путь, в котором содержится `/cmd/`,
они сразу понимают, что получили. Подобным образом устроен репозиторий сборочной утилиты [gb](https://github.com/constabulary/gb).

## Зависимости

Для управления зависимостиями используется утилита [govendor](https://github.com/kardianos/govendor).
Ее выбор вместо [godep](https://github.com/tools/godep) обусловлен более богатым функционалом и отсутсвием проблем при
работе с go 1.6+.

Обязательно нужно ознакомиться с инструкцией по работе с `govendor`.

## Тесты

```sh
$ make test
```

## CI

CI происходит в gitlab ci, все настройки в файле `.gitlab-ci.yml`.

Документация по синтаксису [.gitlab-ci.yml](http://doc.gitlab.com/ce/ci/yaml/README.html)

Используется общий docker runner.

Так же есть [документация](https://gitlab.com/gitlab-org/gitlab-ce/blob/76109d754e167e05db7897f6b89a36b2fadffc65/doc/ci/examples/test-golang-application.md),
в которой есть пример настройки окружения не в docker контейнере.

Если когда-нибудь для тестов понадобится база данныхб смотрите раздел services [postgres](http://docs.gitlab.com/ce/ci/services/postgres.html).

