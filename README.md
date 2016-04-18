Gonder
======
[![Gitter](https://badges.gitter.im/Join Chat.svg)](https://gitter.im/Supme/gonder?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Mass sender

Tools mass email lists, personalization, logging receipt, opening referrals.

Written on Golang

Dependencies:
* github.com/nfnt/resize
* github.com/go-sql-driver/mysql
* github.com/alyu/configparser
* github.com/eaigner/dkim (not use now)
* github.com/gin-gonic/gin

API
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
    	Example:
    	- Get list recipients: http://host/api/recipients?content=recipients&campaign=4&cmd=get-records&limit=100&offset=0
    	- Get recipient parameters: http://host/api/recipients?content=parameters&recipient=149358&cmd=get-records&limit=100&offset=0
    	- ...

License
-------
Distributed under MIT License, please see license file in code for more details.