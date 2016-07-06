Gonder
======
[![Gitter](https://badges.gitter.im/Join Chat.svg)](https://gitter.im/Supme/gonder?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

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

### API
-------
* Groups
	- Example:
	- Get groups: http://host/api/groups?cmd=get-records&limit=100&offset=0
	- Rename groups: http://host/api/groups?cmd=save-records&selected[]=20&limit=100&offset=0&changes[0][recid]=1&changes[0][name]=Test+1&changes[1][recid]=2&changes[1][name]=Test+2
* Campaigns from group
	- Example:
	- Get campaigns: http://host/api/campaigns?group=2&cmd=get-records&limit=100&offset=0
	- Rename campaign: http://host/api/campaigns?cmd=save-records&selected[]=6&limit=100&offset=0&changes[0][recid]=6&changes[0][name]=Test+campaign
* Campaign data
	- Example:
	- Get data: http://host/api/campaign?cmd=get-data&recid=4
	- Save data: ...
* Profiles
	- Example:
	- Get list http://host/api/profiles?cmd=get-list
	- ...
* Recipients from campaign
    - Example:
    - Get list recipients: http://host/api/recipients?content=recipients&campaign=4&cmd=get-records&limit=100&offset=0
    - Get recipient parameters: http://host/api/recipients?content=parameters&recipient=149358&cmd=get-records&limit=100&offset=0
    - ToDo in Wiki

License
-------
Distributed under MIT License, please see license file in code for more details.
