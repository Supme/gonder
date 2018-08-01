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

## License
Distributed under MIT License, please see license file in code for more details.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FSupme%2Fgonder.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FSupme%2Fgonder?ref=badge_large)