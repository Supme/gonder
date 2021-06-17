
Gonder API examples
===

Target ```https://youhost.tld:apiport/```

Use basic authorization for request.

All request in /api route send as parameter "request". For example:

```/api/groups?request={"cmd":"get","offset":0,"limit":100}```

or send JSON in body with Content-Type: application/json

Method POST or GET

Standard error response example:
```json
{"status": "error", "message": "Something error text"}
``` 
and success response example:
```json
{"status": "success", "message": ""}
``` 


## Contents
- [Groups](#Groups)
- [Campaigns](#Campaigns)
- [Campaign](#Campaign)
- [Helpers](#Helpers)
- [Recipients](#Recipients)
- [Reports](#Reports)
- [AdvancedReports](#AdvancedReports)

### Groups
___

Target URI: ```/api/groups```

<details> 

<summary>Get groups list</summary>

```json
{
  "cmd":"get",
  "offset":0,
  "limit":100,
  "sort":
    [
      {"field":"recid","direction":"DESC"}
    ],
    "search":
      [
        {
        "field":"recid",
        "operator":"begins",
        "value":"12"
        }
      ]
}
```
response:
```json 
{
 "total":2,
 "records":
   [
     {"recid":2,"name":"Group 2"},
     {"recid":1,"name":"Group 1"}
   ]
 }
```

</details>


<details> 

<summary>Add group</summary>

```json
{
 "cmd":"add",
  "name": "You name"
}
```

create new group with name "New group" and return added ID in response:

```json
{
  "recid":3
}
```

</details>

	
<details> 

<summary>Save groups name</summary>	
	 
```json
{
  "cmd":"save",
  "offset":0,
  "limit":100,
  "changes":
    [
      {
        "recid":3,
        "name":"New name for group 3"
      },
      {
        "recid":5,
        "name":"New name for group 5"
      }
    ]
}
```

response:

```json
{
  "total":3,
  "records":
    [
      {"recid":3, "name":"New name for group 3"},
      {"recid":5,"name":"New name for group 5"}
    ]
}
```

</details>


### Campaigns
___

Target URI: ```/api/campaigns```

<details> 

<summary>Get campaigns list</summary>

```json
{
  "cmd":"get",
  "id":3,
  "limit":100,
  "offset":0,
    "search":
      [
        {
        "field":"name",
        "operator":"begins",
        "value":"Best"
        }
      ],
  "sort":
    [
      {"field":"name","direction":"ASC"}
    ]
}
```
response:
```json
{
  "total":2,
  "records":
    [
      {"recid":16,"name":"Best A campaign"},
      {"recid":7,"name":"Best B campaign"}
    ]
}
```

</details>


<details> 

<summary>Add campaign</summary>

```json
{
 "cmd":"add",
  "name": "You name"
}
```

create new campaign with name "New campaign" and return added ID in response:

```json
{
  "recid":3
}
```

</details>


<details> 

<summary>Save campaigns name</summary>

```json
{
  "cmd":"save",
  "limit":100,
  "offset":0,
  "changes":
  [
    {"recid":1,"name":"My campaign 1"},
    {"recid":4,"name":"My campaign 4"}
  ]
}
```
response:
```json
{
  "total":2,
  "records":
    [
      {"recid":1,"name":"Campaign 1"},
      {"recid":4,"name":"My campaign 4"}
    ]
}
```

</details>

<details> 

<summary>Clone campaign</summary>

```json
{
  "cmd":"clone",
  "id":23
}
```

clone campaign in new campaign this name "[Clone] Original campaign name" and return added ID in response:

```json
{
  "recid":32,
  "name": "[Clone] Original campaign name"
}
```

</details>


### Campaign
___

Target URI: ```/api/campaign```

<details> 

<summary>Get campaign parameters</summary>

```json
{
  "cmd":"get",
  "id":2
}
```
response: 
```json
{
  "recid":2,
  "name":"My campaign with id 2",
  "profileId":1,
  "subject":"Hello from Gonder",
  "senderId":1,
  "startDate":1479808800,
  "endDate":1480100400,
  "sendUnsubscribe":true,
  "accepted":true,
  "compressHTML": false,
  "templateHTML":"<h1>My cool mail template<h1>",
  "templateText":"My cool mail template"
}
```

</details>


<details> 

<summary>Save campaign parameters</summary>

```json
{
  "cmd":"save",
  "id":2,
  "content":
    {
      "name":"My campaign with id 2",
      "profileId":1,
      "subject":"Hello from Gonder",
      "senderId":1,
      "startDate":1479808800,
      "endDate":1480100400,
      "sendUnsubscribe":true,
      "accepted":true,
      "compressHTML": false,
      "templateHTML":"<h1>My cool mail template<h1>",
      "templateText":"My cool mail template"
    }
}
```
response: 
```json
{
  "recid":2,
  "name":"My campaign with id 2",
  "profileId":1,
  "subject":"Hello from Gonder",
  "senderId":1,
  "startDate":1479808800,
  "endDate":1480100400,
  "sendUnsubscribe":true,
  "accepted":true,
  "compressHTML": false,
  "templateHTML":"<h1>My cool mail template<h1>",
  "templateText":"My cool mail template"
}
```

</details>


<details> 

<summary>Set campaign accept status</summary>

```json
{
  "cmd":"accept",
  "campaign":31,
  "select": true
}
```
response standard json error or success

</details>
 
### Helpers
___

<details> 

<summary>Profile list</summary>

Target URI: ```/api/profilelist```
```json
{
  "cmd":"get"
}
```
response
```json
[
  {"id":1,"text":"Default"},
  {"id":2,"text":"Second IP"},
  {"id":3,"text":"Group from all IP"}
]
```

</details>


<details> 

<summary>Sender list</summary>

Target URI: ```/api/profilelist```
```json
{
  "cmd":"get",
  "id":2
}
```
response
```json
[
  {"id":1,"text":"Gonder (gonder@email.tld)"},
  {"id":4,"text":"Go Sender (gonder@email.tld)"}
]
```

</details>


### Recipients
___

Target URI: ```/api/recipients```

<details> 

<summary>Get recipients list</summary>

```json
{
  "cmd":"get",
  "campaign":1,
  "limit":100,
  "offset":0,
  "sort":
    [
      {"field":"email","direction":"asc"}
    ],
  "search":
    [
      {
        "field":"email",
        "operator":"contains",
        "value":"mail.ru"
      },
      {
        "field":"result",
        "operator":"is",
        "value":"Ok"
      }
    ],
    "searchLogic":"AND"
}
```
response:
```json
{
  "total":2,
  "records":
    [
      {
        "recid":2,
        "name":"Bob",
        "email":"bob@email.com",
        "open": true,
        "result":"Ok"
      },{
         "recid":1,
         "name":"Alice",
         "email":"alice@email.com",
         "open": false,
         "result":""
      }
    ]
}
```

</details>


<details> 

<summary>Get recipient parameters</summary>

```json
{
  "cmd":"get",
  "recipient":2
}
```
response: 
```json
{
 "total":2,
 "records":
   {"Reference":"Bob Sinclair", "Gender":"Man"}
}
```     

</details>


<details> 

<summary>Add recipients to campaign</summary>

```json
{
  "cmd":"add",
  "campaign":2,
  "recipients":
    [
      {
        "name":"Bob",
        "email":"bob@email.tld",
        "params": 
         {
             "Age":"25",
             "Gender":"male"
         }
      },
      {
        "name":"Alice",
        "email":"alice@email.tld",
        "params":
        {
          "Age":"21",
          "Gender":"female"
        }
      }
    ]
}
```
response
```json
{"status": "success", "message": ""}
```
or error
```json
{"status": "error", "message": "Something error text"}
```

</details>


<details> 

<summary>Delete recipients</summary>

```json
{
  "cmd":"delete",
  "ids":
    [1,2,30,40]
}
```
response
```json
{"status": "success", "message": ""}
```
or error
```json
{"status": "error", "message": "Something error text"}
```

</details>

<details>

<summary>Upload recipients file progress</summary>

```json
{
  "cmd":"progress",
  "name":"/tmp/gonder_recipient_load_763792762"
}
```
response progress in percent:
```json
{
  "status": "success",
  "message": 45
}
```
response finish (progress not found)
```json
{
  "status": "error",
  "message": "not found"
}
```

</details>


<details> 

<summary>Clear all recipients from campaign</summary>

```json
{
  "cmd":"clear",
  "campaign":21
}
```
response standard json error or success 

</details>


<details> 

<summary>Mark recipients with result code 4XX (safe bounce) for resend</summary>

```json
{
  "cmd":"resend4xx",
  "campaign":31
}
```
response standard json error or success

</details>


<details> 

<summary>Remove duplicated recipients email list</summary>

```json
{
  "cmd":"deduplicate",
  "campaign":38
}
```
response standard json success with message as count removed recipients or standard error json

</details>


<details> 

<summary>Mark unavaible latest 30 days recipients email (by latest smtp response) list</summary>

```json
{
  "cmd":"unavaible",
  "campaign":22
}
```
response standard json success with message as count marked recipients or standard error json

</details>

### Reports
___

<details> 

<summary>Started campaigns</summary>

request example ```/report/started```

response show id's running campaigns
```json
{"started":["22","43","56"]}
```

</details>

<details> 

<summary>Campaign summary</summary>

request example ```/report/summary?campaign=2318```

response
```json
{
  "Campaign": {
    "Start": 1522945800,
    "name": "My best campaign"
  },
  "OpenMailCount": 4234,
  "OpenWebVersionCount": 41,
  "RecipientJumpCount": 152,
  "SendCount": 26153,
  "SuccessSendCount": 25660,
  "UnsubscribeCount": 3
}
```

</details>


<details> 

<summary>Campaign clicks count</summary>

request example ```/report/clickcount?campaign=2318```

response show count clicks to links
```json
{
  "[Соц.сеть/Facebook]http://www.facebook.com/JaguarRussia/": 68,
  "[Соц.сеть/Instagram]http://instagram.com/jaguarrussia": 32,
  "[Соц.сеть/Twitter]https://twitter.com/JaguarRussia": 26,
  "[Соц.сеть/YouTube]http://www.youtube.com/user/JaguarRussia": 19
}
```

</details>

<details> 

<summary>Campaign recipients list</summary>

request example ```/report/recipients?campaign=2318```

response show count clicks to links
```json
[
  {
    "id":1726190,
    "email":"Alice@mail.tld",
    "name":"Alice",
    "date":1524505276,
    "open":true,
    "status":"Ok"
  },
  {
    "id":1726191,
    "email":"bob@mail.tld",
    "name":"Bob",
    "date":1524505275,
    "open":false,
    "status":"Ok"
  }
]
```

</details>

<details> 

<summary>Recipient clicks</summary>

request example ```/report/clicks?recipient=1726190```

response show count clicks to links
```json
[
  {
    "url":"web_version",
    "date":1524505287
  },
  {
    "url":"open_trace",
    "date":1524505288
  },
  {
    "url": "[Pikabu]https://pikabu.ru/",
    "date":1524505355
  }
]
```

</details>

<details> 

<summary>Campaign or group unsubscribed</summary>

request example show unsubscribe from campaign

```/report/unsubscribed?campaign=3317```

or from group

```/report/unsubscribed?group=7```

```json
[
  {
    "email": "Alice@mail.tld",
    "date": 1545908606,
    "extra": [
          {
            "Unsubscribed": "from header link"
          }
        ]
  },
  {
    "email": "bob@mail.tld",
    "date": 1545908621
  },
  {
    "email": "stive@mail.tld",
    "date": 1545908632,
    "extra": [
          {
            "why": "No time to read"
          }
        ]
  },
  {
    "email": "ivan@mail.tld",
    "date": 1545908634
  }
]
```

</details>

<details> 

<summary>Recipient questions result</summary>

request example ```/report/question?campaign=53```

response show count clicks to links
```json
[
  {
    "recipient_id":1731227,
    "email":"bob@mail.tld",
    "at":1554140897,
    "data":{
      "v2":"emailmarketing",
      "v4":"Push-уведомления"
    }
  },
  {
    "recipient_id":1731227,
    "email":"stive@mail.tld",
    "at":1554141049,
    "data":{
      "v2":"emailmarketing",
      "v3":"powerBI",
      "v4":"Push-уведомления"
    }
  },
  {
    "recipient_id":1731227,
    "email":"ivan@mail.tld",
    "at":1554141065,
    "data":{
      "v2":"emailmarketing",
      "v3":"powerBI",
      "v4":"Push-уведомления"
    }
  }
]
```
</details>



### AdvancedReports
___

All responses can be in json (default) or csv format. For csv use "format=csv" parameter by request.

<details> 

<summary>List campaigns in group</summary>

request example ```/report/group?id=3&type=campaigns```

response:
```json
{
 "status":"ok",
 "message":[
  {
   "id":1,
   "name":"Test campaign",
   "subject":"⚡Gonder test",
   "start":"2019-10-08 11:00:00",
   "end":"2019-11-01 18:00:00"
  },
  {
   "id":2,
   "name":"Test campaign 2",
   "subject":"Test 2",
   "start":"2019-04-22 14:30:00",
   "end":"2019-10-24 15:00:00"
  }
 ]
}
```
</details>

<details> 

<summary>Unsubscribed recipient in group</summary>

request example ```/report/group?id=3&type=unsubscribed```

response:
```json
{
 "status":"ok",
 "message":[
  {
   "campaign_id":1,
   "email":"alice@domain.tld",
   "at":"2019-10-03 12:13:49",
   "data":null
  },
  {
   "campaign_id":2,
   "email":"bob@domain.tld",
   "at":"2019-10-03 13:28:12",
   "data":{"Unsubscribed":"from header link"}
  }
 ]
}
```
</details>

<details> 

<summary>Recipients with status in campaign</summary>

request example ```/report/campaign?id=1&type=recipients```

response:
```json
{
 "status":"ok",
 "message":[
  {
   "id":127,
   "email":"alice@domain.tld",
   "name":"Alice",
   "at":"2019-10-24 14:43:01",
   "status":"Ok",
   "open":false,
   "data":{"Age":"27","Gender":"f"}
  },
  {
   "id":128,
   "email":"bob@domain.tld",
   "name":"Bob",
   "at":"2019-10-24 14:43:02",
   "status":"Ok",
   "open":true,
   "data":{"Age":"29","Gender":"m"}
  }
 ]
}
```
</details>

<details> 

<summary>Recipients clicks (jump to url) in campaign</summary>

request example ```/report/campaign?id=3&type=clicks```

response:
```json
{
 "status":"ok",
 "message":[
  {
   "id":21,
   "email":"alice@domain.tld",
   "at":"2019-06-17 13:54:37",
   "url":"web_version"
  },
  {
   "id":21,
   "email":"bob@domain.tld",
   "at":"2019-06-17 14:24:23",
   "url":"[Gonder git] https://github.com/Supme/gonder/"
  }
 ]
}
```
</details>

<details>

<summary>Unsubscribed recipients clicks in campaign</summary>

request example ```/report/campaign?id=1&type=unsubscribed```

response:
```json
{
 "status":"ok",
 "message":[
  {
   "email":"alice@domain.tld",
   "at":"2019-10-03 12:13:49",
   "data":null
  },
  {
   "email":"bob@domain.tld",
   "at":"2019-10-03 13:28:12",
   "data":{"Unsubscribed":"from header link"}
  }
 ]
}
```
</details>

<details>

<summary>Recipients answer the questions in campaign</summary>

request example ```/report/campaign?id=1&type=question```

response:
```json
{
 "status":"ok",
 "message":[
  {
   "id":5,
   "email":"alice@domain.tld",
   "at":"2019-04-05 10:28:17",
   "data":{"v2":"emailmarketing","v3":"powerBI"}
  },
  {
   "id":9,
   "email":"bob@domain.tld",
   "at":"2019-04-09 09:45:34",
   "data":{"v2":"emailmarketing","v4":"Push-уведомления"}
  }
 ]
}
```
</details>

<details>

<summary>Recipients user agent (email client and web browser)</summary>

request example ```/report/campaign?id=2&type=useragent```

response:
```json
{
"status":"ok",
 "message":[
  {
   "id":44,
   "email":"alice@domain.tld",
   "name":"Alice",
   "client":{
    "ip":"77.88.31.235",
    "is_mobile":false,
    "is_bot":true,
    "platform":"",
    "os":"",
    "engine_name":"",
    "engine_version":"",
    "browser_name":"YandexImageResizer",
    "browser_version":"2.0"
    },
   "browser":null
  },
  {
   "id":45,
   "email":"bob@domain.tld",
   "name":"Bob",
     "client":null,
     "browser":{
      "ip":"1.2.3.4",
      "is_mobile":false,
      "is_bot":false,
      "platform":"Windows",
      "os":"Windows 7",
      "engine_name":"AppleWebKit",
      "engine_version":"537.36",
      "browser_name":"Chrome",
      "browser_version":"76.0.3809.100"
     }
  }
 ]
}
```
</details>
