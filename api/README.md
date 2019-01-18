
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
 "cmd":"add"
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

<summary>Save groups</summary>	
	 
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
  "template":"<h1>My cool mail template<h1>"
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
      "template":"<h1>My cool mail template<h1>"
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
  "template":"<h1>My cool mail template<h1>"
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
   [
     {"key":"Reference","value":"Bob Sinclair"},
     {"key":"Gender","value":"Man"}
   ]
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
         [
           {
             "key":"Age",
             "value":"25"
           },
           {
             "key":"Gender",
             "value":"male"
           }
         ]
      },
      {
        "name":"Alice",
        "email":"alice@email.tld",
        "params": 
          [
            {
              "key":"Age",
              "value":"21"
            },
            {
              "key":"Gender",
              "value":"female"
            }
          ]
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

<summary>Upload recipients file list</summary>

```json
{
  "cmd":"upload",
  "campaign":2,
  "fileName":"my_subscribers.xlsx",
  "fileContent":"base64 coded file content"
}
```
response:
```json
{
  "status": "success",
  "message": "/tmp/gonder_recipient_load_763792762"
}
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

request example ```/report/status```

response show id's running campaigns
```json
{"started":["22","43","56"]}
```

</details>

<details> 

<summary>Campaign summary</summary>

request example ```/report?campaign=2318```

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

<summary>Campaign link clicks</summary>

request example ```/report/jump?campaign=2318```

response show count jumping to links
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