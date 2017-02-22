Gonder
======

Target https://youhost:apiport/api/... basic authorization

* Groups groups?request={json}
	- {"cmd":"get","limit":100,"offset":0}
	response:
	 {"total":2,"records":[{"recid":2,"name":"Group 2"},{"recid":1,"name":"Group 1"}]}
	- {"cmd":"add"}
	response:
	 {"recid":3,"name":"New group"}
	- {"cmd":"save","limit":100,"offset":0,"changes":[{"recid":3,"name":"Group 3"}]}
	response:
	 {"total":3,"records":[{"recid":3,"name":"Group 3"},{"recid":2,"name":"Group 2"},{"recid":1,"name":"Group 1"}]}
	
* Campaigns campaigns?request={json}
    - {"cmd":"get","id":3,"limit":100,"offset":0}
    response:
     {"total":2,"records":[{"recid":2,"name":"Campaign 2"},{"recid":1,"name":"Campaign 1"}]}
    - {"cmd":"save","limit":100,"offset":0,"changes":[{"recid":1,"name":"My campaign 1"}]}
    response:
     {"total":2,"records":[{"recid":2,"name":"Campaign 2"},{"recid":1,"name":"My campaign 1"}]}
    
* Campaign campaign?request={json}
    - {"cmd":"get","id":2}
    response: 
     {"recid":1,"name":"My campaign","profileId":1,"subject":"Hello from Gonder","senderId":1,"startDate":1479808800,"endDate":1480100400,"sendUnsubscribe":true,"accepted":true,"template":"My cool mail template"}
    
* Recipients recipients?request={json}
    - {"cmd":"get","campaign":1,"limit":100,"offset":0}
   response:
    {"total":2,"records":[{"recid":2,"name":"Bob","email":"bob@email.com","result":"Ok"},{"recid":1,"name":"Alice","email":"alice@email.com","result":""}]}
    - {"cmd":"get","limit":100,"offset":0,"recipient":2}
    response: 
     {"total":2,"records":[{"key":"Reference","value":"Bob Sinclair"}, {"key":"Gender","value":"Man"}]}
    - {"cmd":"upload","campaign":2,"fileName":"gosender.xlsx","fileContent":"base64 coded file content"}