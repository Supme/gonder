
$(document).ready(
    function () {
        $.ajaxSetup({
            cache: true
        });
        w2utils.locale('assets/w2ui/locale/ru-ru.json?{{.version}}');
        w2utils.locale('assets/panel/locale/ru-ru.json?{{.version}}');
        $.getScript('assets/panel/layout.js?{{.version}}', function() {
            $.getScript('assets/panel/template.js?{{.version}}', function () {


                $.getScript('assets/panel/group.js?{{.version}}', function () {
                    $.getScript('assets/panel/sender.js?{{.version}}');
                    $.getScript('assets/panel/campaign.js?{{.version}}');
                });
                $.getScript('assets/panel/recipient.js?{{.version}}');
                $.getScript('assets/panel/profile.js?{{.version}}');
                $.getScript('assets/panel/users.js?{{.version}}');
            });
        });
    }
);

var ws = null;
var refreshNum;
function startStatusLog(file){
    var logMaxRows = 200;
    var statusData = [];
    var statusLog = $("#statusLog");
    var wsuri = window.location.protocol=="https:"?"wss://":"ws://";
    wsuri += window.location.hostname;
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