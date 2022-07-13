// --- Recipients table ---
$('#campaignRecipient').w2grid({
    name: 'recipient',
    show: {
        toolbar: true,
        footer: true
    },
    multiSearch: true,
    searches: [
        { field: 'recid', label: w2utils.lang('Id'), type: 'int' },
        { field: 'email', label: w2utils.lang('Email'), type: 'text' },
        { field: 'name', label: w2utils.lang('Name'), type: 'text' },
        { field: 'result', label: w2utils.lang('Result'), type: 'text' }
    ],
    columns: [
        { field: 'recid', text: w2utils.lang('Id'), sortable: true, size: '100px', resizable: true,
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
                        $.each(data.records, function (key, val) {
                           table += '<tr><td>' + key + '</td><td>' + val + '</td></tr>';
                        });
                    });
                    return table;
                }
            }
        },
        { field: 'email', text: w2utils.lang('Email'), sortable: true, size: '150px', resizable: true },
        { field: 'name', text: w2utils.lang('Name'), sortable: true, size: '150px', resizable: true },
        { field: 'open', text: w2utils.lang('Opened'), sortable: true, size: '65px', resizable: false, render: 'toggle', style: 'text-align: center' },
        { field: 'result', text: w2utils.lang('Result'), sortable: true, size: '100%', resizable: true }
    ],
    multiSelect: true,
    method: 'GET',
    postData: { cmd:"get" },

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
            {id: 'add', type: 'button', text: w2utils.lang('Add New'), tooltip: w2utils.lang("Add recipient"), icon: 'w2ui-icon-plus'},
            {id: 'delete', type: 'button', text: w2utils.lang('Delete'), tooltip: w2utils.lang("Delete selected recipients"), icon: 'w2ui-icon-cross'},
            {id: 'csv', type: 'button', text: w2utils.lang('CSV'), tooltip: w2utils.lang("Get this as csv file"), icon: 'w2ui-icon-columns'}
        ],
        onClick: function (event) {
            switch(event.target) {
                case "csv":
                    var url = '/report/file/recipients?' +
                        'campaign=' + w2ui.campaign.getSelection()[0] + '&' +
                        'params=' + JSON.stringify({
                            sort: w2ui['recipient'].sortData,
                            search: w2ui['recipient'].searchData,
                            searchLogic: w2ui.recipient.last.logic
                        });
                    loadLink(url);
                    break;

                case "add":
                    addRecipient();
                    break;

                case "delete":
                    deleteRecipients();
                    break;
            }
        }
    }

});
// --- /Recipients table ---

// --- Delete recipient ---
function deleteRecipients() {
    w2confirm(w2utils.lang('Are you sure you want to delete selected records?'))
        .yes(() => {
            $.ajax({
                url: "api/recipients",
                type: "GET",
                dataType: "json",
                data: {"request": JSON.stringify({"cmd": "delete", "ids": w2ui.recipient.getSelection()})}
            }).done(function(data) {
                if (data['status'] === 'error') {
                    w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                }
                w2ui['recipient'].reload();
            });
        });
}
// --- /Delete recipient ---

// --- Add recipient ---
function addRecipient() {
    w2popup.open({
        name: "addRecipientPopup",
        width   : 400,
        height  : 480,
        title   : w2utils.lang("Add recipient"),
        body    : '<div id="addRecipient" style="width: 100%; height: 100%;"></div>',
        onOpen  : function (event) {
            event.onComplete = function () {
                $('#addRecipient').w2render('addRecipient');
            }
        },
        onClose : function (event) {
            w2ui.addRecipient.clear();
        }
    });
    $('#addRecipient').w2form({
        name: 'addRecipient',
        fields : [
            { field: 'email', html: { label: 'Email' }, type: 'email' },
            { field: 'name', html: { label: 'Name' }, type: 'text' },
            { field: 'params', type: 'map',
                html: {
                    label: 'Parameters',
                    key: {
                        attr: 'placeholder="key" style="width: 80px"',
                        text: ' = '
                    },
                    value: {
                        attr: 'placeholder="value" style="width: 100px"',
                    }
                }
            }
        ],
        actions: {
            Reset: function () {
                this.clear();
            },
            Save: function () {
                if (w2ui.addRecipient.validate().length == 0) {
                    let rec = this.getCleanRecord();
                    $.ajax({
                        url: "api/recipients",
                        type: "GET",
                        dataType: "json",
                        data: {"request": JSON.stringify({"cmd": "add", "campaign": parseInt($('#campaignId').val()), "recipients": [{"email": rec.email, "name": rec.name, "params": rec.params}]})}
                    }).done(function(data) {
                        if (data['status'] === 'error') {
                            w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                        }
                        w2ui.recipient.reload();
                    });
                    w2popup.close();
                }
            }
        }
    });
}
// --- /Add recipient ---

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

// --- Recipient unavailable ---
$("#recipientUnavailable").html(w2utils.lang('Mark unavailable'));
$('#recipientUnavailable').click(
    function () {
        w2confirm(w2utils.lang('Mark unavailable recent time recipients') + '?', function (btn) {
            if (btn === 'Yes') {
                w2ui.layout.lock('main', w2utils.lang('Marking unavailable...'), true);
                $.ajax({
                    url: "api/recipients",
                    type: "GET",
                    data: {"request": JSON.stringify({"cmd": "unavailable", "campaign": parseInt($('#campaignId').val()), "interval": parseInt($('#recipientUnavailableDay').val())}),
                        dataType: "json"
                    }
                }).done(function(data) {
                    if (data['status'] !== 'success') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    } else {
                        w2alert(data["message"] + " " + w2utils.lang("marked as unavailable"));
                    }
                    w2ui['recipient'].reload();
                    w2ui.layout.unlock('main');
                })
            }
        })

    }
);
// --- Recipient unavailable ---