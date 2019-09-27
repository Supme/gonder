var ws = null;
var refreshNum;
function startStatusLog(file){
    var logMaxRows = 200;
    var statusData = [];
    var statusLog = $("#statusLog");
    var wsuri = "wss://" + window.location.hostname;
    wsuri += window.location.port==""?"":":"+window.location.port;
    wsuri += "/status/ws/" + file;

    if (ws == null) {
        ws = new WebSocket(wsuri);
    } else {
        ws.close();
        clearInterval(refreshNum);
        setTimeout(function(){}, 2000);
        ws = new WebSocket(wsuri);
    }

    statusLog.empty();

    ws.onopen = function() {
        console.log("websocket connected to " + wsuri);
    };

    ws.onclose = function(e) {
        console.log("websocket connection to " + wsuri + " closed (" + e.code + ") clean = " + e.wasClean);
    };

    ws.onerror = function (e) {
        console.log("websocket connection error " + e);
    };

    refreshNum = setInterval(function () {
        if (statusLog.is(':visible') && (statusLog.prop("scrollTop") == statusLog.prop("scrollHeight")-statusLog.prop("clientHeight"))) {
            statusLog.empty();
            statusData.forEach(function(entry) {
                statusLog.append(entry+"<br/>");
            });
            statusLog.scrollTop(statusLog.prop("scrollHeight"));

        }
    }, logMaxRows);
    ws.onmessage = function(e) {
        statusData.push(e.data);
        if (statusData.length > logMaxRows) {
            statusData = statusData.slice(statusData.length-logMaxRows, statusData.length);
        }
    };
}