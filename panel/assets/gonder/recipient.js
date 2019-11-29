// --- Recipients table ---
$('#campaignRecipient').w2grid({
    name: 'recipient',
    show: {
        toolbar: true,
        footer: true
    },
    multiSearch: true,
    searches: [
        { field: 'recid', caption: w2utils.lang('Id'), type: 'int' },
        { field: 'email', caption: w2utils.lang('Email'), type: 'text' },
        { field: 'name', caption: w2utils.lang('Name'), type: 'text' },
        { field: 'result', caption: w2utils.lang('Result'), type: 'text' }
    ],
    columns: [
        { field: 'recid', caption: w2utils.lang('Id'), sortable: true, size: '100px', resizable: true,
            info: {
                render : function (rec) {
                    var table;
                    $.ajax({
                        type: "GET",
                        async: false,
                        url: '/api/recipients',
                        data: {"request": JSON.stringify({"cmd": "get", "recipient": rec.recid})},
                        dataType: "json"
                    }).done(function(data) {
                        table = '<table>';
                        $.each(data.records, function (i, val) {
                           table += '<tr><td>' + val["key"] + '</td><td>' + val["value"] + '</td></tr>';
                        });
                    });
                    return table;
                }
            }
        },
        { field: 'email', caption: w2utils.lang('Email'), sortable: true, size: '150px', resizable: true },
        { field: 'name', caption: w2utils.lang('Name'), sortable: true, size: '150px', resizable: true },
        { field: 'open', caption: w2utils.lang('Opened'), sortable: true, size: '65px', resizable: false, render: 'toggle', style: 'text-align: center' },
        { field: 'result', caption: w2utils.lang('Result'), sortable: true, size: '100%', resizable: true }
    ],
    multiSelect: false,
    method: 'GET',

    onDblClick: function (event) {
        preview = window.open(
            "/preview?id="+event.recid,
            "Preview mail"+event.recid,
            "width=800,height=600,resizable=yes,scrollbars=yes,status=yes"
        );
        preview.focus();
    },

    toolbar: {
        items: [
            {id: 'csv', type: 'button', caption: w2utils.lang('CSV'), tooltip: w2utils.lang("Get this as csv file"), icon: 'w2ui-icon-columns'}
        ],
        onClick: function (event) {
            if (event.target === "csv") {
                var url = '/report/file/recipients?' +
                    'campaign=' + w2ui.campaign.getSelection()[0] + '&' +
                    'params=' + JSON.stringify({
                        sort: w2ui['recipient'].sortData,
                        search: w2ui['recipient'].searchData,
                        searchLogic: w2ui.recipient.last.logic
                    });
                loadLink(url);
            }
        }
    }

});
// --- /Recipients table ---

// --- Recipient upload ---
$('#recipientUploadFile').w2field('file', {max: 1});
$("#recipientUploadButton").html(w2utils.lang('Upload'));
$('#recipientUploadButton').click(
    function () {
        if (w2ui['toolbar'].get('acceptSend').checked) {
            w2alert(w2utils.lang('Cannot add recipients to an accepted campaign.'), w2utils.lang('Error'));
            return
        }
        if ($('#recipientUploadFile').data("selected").length === 0) {
            w2alert(w2utils.lang('No one file selected.'), w2utils.lang('Error'));
        } else {
            w2ui.layout.lock('main', "<span id='uploadProgress'>"+ w2utils.lang("Uploading 0%") + "</span>", true);
            $.each($('#recipientUploadFile').data('selected'), function(index, file){
                $.ajax({
                    url: "/recipient/upload",
                    type: "POST",
                    dataType: "json",
                    data: {
                            "id": parseInt($('#campaignId').val()),
                            "name": file.name,
                            "content": file.content
                        }
                }).done(function(data) {
                    if (data['status'] === 'error') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                        $('#recipientUploadFile').w2field('file', {max: 1});
                        w2ui['recipient'].reload();
                        w2ui.layout.unlock('main');
                    } else {
                        var finish = false;
                        prs = setInterval(function(){
                            $.ajax({
                                url: "api/recipients",
                                type: "GET",
                                dataType: "json",
                                data: {
                                    "request": JSON.stringify({
                                        "cmd": "progress",
                                        "name": data["message"]
                                    })
                                }
                            }).done(function(req) {
                                if (req["status"] === "success") {
                                    $('#uploadProgress').text("Uploading: " + req["message"] + "%");
                                } else {
                                    finish = true;
                                    console.log(req["status"]);
                                }
                                if (finish) {
                                    clearInterval(prs);
                                    $('#recipientUploadFile').w2field('file', {max: 1});
                                    w2ui['recipient'].reload();
                                    w2ui.layout.unlock('main');
                                };
                            });
                        }, 500);
                    }
                });
            });
        }
    }
);
// --- /Recipient upload ---

// --- Recipient delete all ---
$("#recipientClearButton").html(w2utils.lang('Clear'));
$("#recipientClearButton").click(
    function () {
        w2confirm(w2utils.lang('Delete all recipients from campaign?'), function (btn) {
            if (btn === 'Yes') {
                w2ui.layout.lock('main', w2utils.lang('Deleting...'), true);
                $.ajax({
                    url: "api/recipients",
                    type: "GET",
                    dataType: "json",
                    data: {"request": JSON.stringify({"cmd": "clear", "campaign": parseInt($('#campaignId').val())})}
                }).done(function(data) {
                    if (data['status'] === 'error') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    }
                    w2ui['recipient'].reload();
                    w2ui.layout.unlock('main');
                });
            }
        })
    }
);
// --- /Recipient delete all ---

// --- Recipient resend ---
$("#recipientResend").html(w2utils.lang('Resend by 4xx code'));
$('#recipientResend').click(
    function () {
        w2confirm(w2utils.lang('Resend by 4xx code') + '?', function (btn) {
            if (btn === 'Yes') {
                w2ui.layout.lock('main', w2utils.lang('Update...'), true);
                $.ajax({
                    url: "api/recipients",
                    type: "GET",
                    data: {"request": JSON.stringify({"cmd": "resend4xx", "campaign": parseInt($('#campaignId').val())}),
                    dataType: "json"
                   }
                }).done(function(data) {
                    if (data['status'] === 'error') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    }
                    w2ui['recipient'].reload();
                    w2ui.layout.unlock('main');
                })
            }
        })

    }
);
// --- Recipient resend ---

// --- Recipient deduplicate---
$("#recipientDeduplicate").html(w2utils.lang('Deduplicate'));
$('#recipientDeduplicate').click(
    function () {
        w2confirm(w2utils.lang('Deduplicate recipients') + '?', function (btn) {
            if (btn === 'Yes') {
                w2ui.layout.lock('main', w2utils.lang('Deduplicating...'), true);
                $.ajax({
                    url: "api/recipients",
                    type: "GET",
                    data: {"request": JSON.stringify({"cmd": "deduplicate", "campaign": parseInt($('#campaignId').val())}),
                        dataType: "json"
                    }
                }).done(function(data) {
                    if (data['status'] !== 'success') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    } else {
                        w2alert(data["message"] + " " + w2utils.lang("duplicates removed"));
                    }
                    w2ui['recipient'].reload();
                    w2ui.layout.unlock('main');
                })
            }
        })

    }
);
// --- Recipient deduplicate ---

// --- Recipient unavaible ---
$("#recipientUnavaible").html(w2utils.lang('Mark unavaible'));
$('#recipientUnavaible').click(
    function () {
        w2confirm(w2utils.lang('Mark unavaible recent time recipients') + '?', function (btn) {
            if (btn === 'Yes') {
                w2ui.layout.lock('main', w2utils.lang('Marking unavaible...'), true);
                $.ajax({
                    url: "api/recipients",
                    type: "GET",
                    data: {"request": JSON.stringify({"cmd": "unavaible", "campaign": parseInt($('#campaignId').val())}),
                        dataType: "json"
                    }
                }).done(function(data) {
                    if (data['status'] !== 'success') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    } else {
                        w2alert(data["message"] + " " + w2utils.lang("marked as unavaible"));
                    }
                    w2ui['recipient'].reload();
                    w2ui.layout.unlock('main');
                })
            }
        })

    }
);
// --- Recipient unavaible ---