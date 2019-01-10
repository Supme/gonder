
Gonder API examples
==

Target ```https://youhost.tld:apiport/```

Use basic authorization for request.

All request in /api route send as parameter "request". For example:

```/api/groups?request={"cmd":"get","offset":0,"limit":100}```

or send JSON in body with Content-Type: application/json

Error response example:
```json
{"status": "error", "message": "Something error text"}
``` 

- [groups](#Groups)
- [campaigns](#Campaigns)
- [campaign](#Campaign)


#### Groups

Target URI: ```/api/groups```

- Get groups list
```json
{
  "cmd":"get",
  "offset":0,
  "limit":100
}
```
 example response:
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
- Add group

```json{
 "cmd":"add"
}```

create new group with name "New group" and return added ID in response:

```json
{
  "recid":3
}
```
	
- Save groups	
	 
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

Target URI: ```/api/campaigns```

- Get campaigns list
```json
{
  "cmd":"get",
  "id":3,
  "limit":100,
  "offset":0
}
```
response:
```json
{
  "total":2,
  "records":
    [
      {"recid":2,"name":"Campaign 2"},
      {"recid":1,"name":"Campaign 1"}
    ]
}
```
- Save campaigns name
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

##### Campaign
Target URI: ```/api/campaigns```
- Get campaign parameters
```json
{"cmd":"get","id":2}
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
  "template":"<h1>My cool mail template<h1>"
}
```
    
#####Recipients
Target URI: ```/api/recipients```
```json
{
  "cmd":"get",
  "campaign":1,
  "limit":100,
  "offset":0
}
```
   response:
    {"total":2,"records":[{"recid":2,"name":"Bob","email":"bob@email.com","result":"Ok"},{"recid":1,"name":"Alice","email":"alice@email.com","result":""}]}
    - {"cmd":"get","limit":100,"offset":0,"recipient":2}
    response: 
     {"total":2,"records":[{"key":"Reference","value":"Bob Sinclair"}, {"key":"Gender","value":"Man"}]}
    - {"cmd":"upload","campaign":2,"fileName":"gosender.xlsx","fileContent":"base64 coded file content"}