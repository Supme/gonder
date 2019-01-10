
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

- [groups](#Groups)
- [campaigns](#Campaigns)
- [campaign](#Campaign)


### Groups
___

Target URI: ```/api/groups```

##### Get groups list

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
##### Add group

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
	
##### Save groups	
	 
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

### Campaigns
___

Target URI: ```/api/campaigns```

##### Get campaigns list
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
##### Save campaigns name
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

### Campaign
___

Target URI: ```/api/campaigns```

##### Get campaign parameters
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

##### Save campaign parameters
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

##### Set campaign accept status
```json
{
  "cmd":"accept",
  "campaign":31,
  "select": true
}
```
response standard json error or success
 

### Recipients
___

Target URI: ```/api/recipients```

##### Get recipients list
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

##### Get recipient parameters
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

##### Add recipients to campaign
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

##### Upload recipients file list
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

##### Upload recipients file progress
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

##### Clear all recipients from campaign
```json
{
  "cmd":"clear",
  "campaign":21
}
```
response standard json error or success 


##### Mark recipients with result code 4XX (safe bounce) for resend
```json
{
  "cmd":"resend4xx",
  "campaign":31
}
```
response standard json error or success


##### Remove duplicated recipients email list
```json
{
  "cmd":"deduplicate",
  "campaign":38
}
```
response standard json success with message as count removed recipients or standard error json


##### Mark unavaible latest 30 days recipients email (by response smtp response) list
```json
{
  "cmd":"unavaible",
  "campaign":22
}
```
response standard json success with message as count marked recipients or standard error json
