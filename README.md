# funpayUpdate

Небольшой демон на Go, который периодически «поднимает» лоты на
[FunPay](https://funpay.com) — логинится через VK-гейт в headless Chrome и
жмёт кнопку поднятия в нужной категории.

## Требования

- Go 1.24+
- Установленный Chrome / Chromium
- `chromedriver`, совпадающий по версии с браузером
  ([скачать](https://googlechromelabs.github.io/chrome-for-testing/))

## Сборка

```sh
go build -o bin/funpayupdater ./cmd/funpayupdater
```

## Запуск

Пароль читается **только** из переменной окружения — флаги командной строки
видны всем процессам в системе.

```sh
export FUNPAY_USERNAME='...'
export FUNPAY_PASSWORD='...'
./bin/funpayupdater
```

Разовый прогон для проверки настроек:

```sh
./bin/funpayupdater -once
```

### Настройки

| Флаг / переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `FUNPAY_PASSWORD` | — | Пароль VK. Обязателен, только через env. |
| `-username` / `FUNPAY_USERNAME` | — | Логин VK. Обязателен. |
| `-lots-url` | `https://funpay.com/lots/1120/trade` | Страница категории с кнопкой поднятия. |
| `-interval` | `4h18m` | Пауза между поднятиями. |
| `-once` | `false` | Поднять один раз и выйти. |
| `-chromedriver` / `CHROMEDRIVER_PATH` | ищется в `PATH` | Путь к chromedriver. |
| `-port` | `5050` | Локальный порт chromedriver. |
| `-screenshot` | `Screenshot.jpg` | Куда сохранять скриншот после поднятия. Пустая строка — отключить. |

## Структура

```
cmd/funpayupdater/   точка входа: флаги, планировщик, graceful shutdown
internal/funpay/     один цикл «залогиниться и поднять»
```

## Оговорка

Селекторы захардкожены XPath'ами по вёрстке FunPay и VK — при любом редизайне
страниц их придётся обновить в `internal/funpay/funpay.go`. Одиночная неудача
логируется, и демон продолжает работать по расписанию; в режиме `-once`
ошибка возвращает ненулевой код выхода.
