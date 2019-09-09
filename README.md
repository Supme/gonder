Gonder
======
[![Go Report Card](https://goreportcard.com/badge/github.com/Supme/gonder)](https://goreportcard.com/report/github.com/Supme/gonder)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FSupme%2Fgonder.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FSupme%2Fgonder?ref=badge_shield)

Mass sender

Tools mass email lists, personalization, logging receipt, opening referrals.

Written on Golang

### Возможности:
* Многопоточная рассылка.
* Профили рассылки (выбор сетевого интерфейса, количества потоков рассылки, паузы между досылками писем по одному и количество попыток доотправок).
* Возможность отправки через SOCKS5.
* Шаблонизатор со всеми вытекающими персонализациями писем.
* Веб версия письма.
* Функционал отписки с возможностью изменения страниц отписки для каждой группы.
* Раздельные группы кампаний, каждая со своими отправителями.
* Статистика кампаний (открытия писем, переходов по ссылкам, отписки).
* Веб панель управления работающая через API.
* Полное разграничение прав доступа по группам и действиям в API/панели.

### Требования:
* MySQL или аналогичная БД.
* Linux, Windows (полное тестирование ведётся на Linux, запуск на Windows только нечасто проверяется).
* Существование и правильное внесение SPF/DKIM/DMARK записей в DNS домена от имени которого ведётся рассылка.
* Существование почтового ящика от имени которого ведутся рассылки.
* Существование и верно заданная прямая и обратная записи IP адреса и её соответствие указанному в профиле рассылки.
* Соблюдение общих требований к честным и легальным рассылкам.

### Run
Use dist_config.ini as example config

Create certificates (use README.md) in "cert" folder.

Get database dump for MySQL/MariaDb github https://raw.githubusercontent.com/Supme/gonder/master/dump.sql

or initialize database command:
```shell script
./gonder -i
```

```shell script
Usage of ./gonder:
  -c  	    Path to config file (default "./dist_config.ini")
  -i	    Initial database
  -iy  	    Initial database without confirm
  -p   	    Path to pid file (default "pid/gonder.pid")
  -restart  Restart daemon
  -start    Start as daemon
  -status   Check daemon status
  -stop     Stop daemon
  -v	Prints version

```

Open in browser ```https://[host]:[api_port][panel_path]```

Default admin user for panel: admin:admin

### Docker
Build:
```shell script
git clone https://github.com/Supme/gonder.git
cd gonder
docker build -t gonder .
```

Or use dockerhub:
```shell script
docker pull supme/gonder
```

Run:
```shell script
docker run -d -i -t --rm --network host --name gonder \
-e GONDER_MAIN_DEFAULT_PROFILE_ID=1 \
-e GONDER_DATABASE_STRING='gonder:gonderpass@tcp(127.0.0.1:3306)/gonderdb' \
-e GONDER_MAILER_SEND=true \
-e GONDER_UTM_DEFAULT_URL='http://localhost:8080' \
-e GONDER_UTM_PORT=8080 \
-e GONDER_API_PORT=7777 \
-e GONDER_API_PANEL_PATH='/panel' \
-e GONDER_API_PANEL_LOCALE='ru-ru' \
-v files:/app/files \
gonder
```

## License
Distributed under MIT License, please see license file in code for more details.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FSupme%2Fgonder.svg?type=small)](https://app.fossa.io/projects/git%2Bgithub.com%2FSupme%2Fgonder?ref=badge_small)