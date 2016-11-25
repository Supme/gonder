// --- Recipients table ---
$('#campaignRecipient').w2grid({
    header: w2utils.lang("Recipients"),
    name: 'recipient',
    show: {
        header: true,
        footer: true
    },
    columns: [
        { field: 'recid', caption: w2utils.lang('Id'), size: '5%' },
        { field: 'email', caption: w2utils.lang('Email'), size: '15%' },
        { field: 'name', caption: w2utils.lang('Name'), size: '20%' },
        { field: 'result', caption: w2utils.lang('Result'), size: '60%' }
    ],
    multiSelect: false,
    method: 'GET',
    onExpand: function (event) {
        if (w2ui.hasOwnProperty('subgrid-' + event.recid)) w2ui['subgrid-' + event.recid].destroy();
        $('#'+ event.box_id).css({ margin: '0px', padding: '0px', width: '100%' }).animate({ height: '100px' }, 100);
        setTimeout(function () {
            $('#'+ event.box_id).w2grid({
                name: 'subgrid-' + event.recid,
                fixedBody: true,
                columns: [
                    { field: 'key', caption: w2utils.lang('Parameter'), size: '30%' },
                    { field: 'value', caption: w2utils.lang('Value'), size: '70%' },
                ],
                multiSelect: false,
                postData: { "recipient": parseInt(event.recid) },
                url: '/api/recipients', //?content=parameters&recipient='+event.recid,
                method: 'GET'
            });
            w2ui['subgrid-' + event.recid].resize();
        }, 300);
    },
    onDblClick: function (event) {
        /*w2popup.load({
         title: 'Mail preview',
         url: '/preview?id='+event.recid,
         showMax: true
         });*/
        preview = window.open(
            "/preview?id="+event.recid,
            "Preview mail"+event.recid,
            "width=800,height=600,resizable=yes,scrollbars=yes,status=yes"
        );
        preview.focus();
    }
});
// --- /Recipients table ---

// --- Recipient upload ---
$('#recipientUploadFile').w2field('file', {max: 1});
$("#recipientUploadButton").html(w2utils.lang('Upload'));
$('#recipientUploadButton').click(
    function () {
        if ($('#recipientUploadFile').data("selected").length == 0) {
            w2alert(w2utils.lang('No one file selected.'), w2utils.lang('Error'));
        } else {
            w2ui.layout.lock('main', w2utils.lang('Uploading...'), true);
            $.each($('#recipientUploadFile').data('selected'), function(index, file){
                $.ajax({
                    url: "api/recipients",
                    type: "POST",
                    dataType: "json",
                    data: {
                        "request": JSON.stringify({
                            "cmd": "upload",
                            "campaign": parseInt($('#campaignId').val()),
                            "fileName": file.name,
                            "fileContent": file.content
                        })
                    }
                }).done(function(data) {
                    if (data['status'] == 'error') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    }
                    $('#recipientUploadFile').w2field('file', {max: 1});
                    w2ui['recipient'].reload();
                    w2ui.layout.unlock('main');
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
            if (btn == 'Yes') {
                w2ui.layout.lock('main', w2utils.lang('Deleting...'), true);
                $.ajax({
                    url: "api/recipients",
                    type: "GET",
                    dataType: "json",
                    data: {"request": JSON.stringify({"cmd": "clear", "campaign": parseInt($('#campaignId').val())})}
                }).done(function(data) {
                    if (data['status'] == 'error') {
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
$("#recipientResend").html(w2utils.lang('Resend by 4x1 code'));
$('#recipientResend').click(
    function () {
        console.log("click resend");
        w2confirm(w2utils.lang('Resend by 421 and 451 code?'), function (btn) {
            if (btn == 'Yes') {
                w2ui.layout.lock('main', w2utils.lang('Deleting...'), true);
                $.ajax({
                    url: "api/recipients",
                    type: "GET",
                    data: {
                        content: "recipients",
                        cmd: "resend4x1",
                        campaign: $('#campaignId').val(),
                    },
                    dataType: "json",
                    data: {
                        "request": JSON.stringify({
                            "cmd": "resend4x1",
                            "campaign": parseInt($('#campaignId').val())
                        })
                    }
                }).done(function(data) {
                    if (data['status'] == 'error') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    }
                }).complete(function() {
                    w2ui['recipient'].reload();
                    w2ui.layout.unlock('main');
                })
            }
        })

    }
);
// --- Recipient resend ---