function refreshSenderList(selectedSender){
    var sender;
    $.ajax({
        type: "GET",
        async: false,
        url: '/api/senderlist',
        dataType: "json",
        data: {"request": JSON.stringify({"cmd": "get", "id": parseInt(w2ui['group'].getSelection()[0])})},
    }).done(function(data) {
        sender = data;
    });
    console.log(sender);
    w2ui['parameter'].set('campaignSenderId', { options: { items: sender } });
    w2ui['parameter'].record['campaignSenderId'] = selectedSender;
    w2ui['parameter'].refresh();
}

// --- Sender emails editor ---
$().w2layout({
    name: 'senderEditor',
    padding: 4,
    panels: [
        { type: 'left', size: '40%', resizable: true, minSize: 300 },
        { type: 'main', minSize: 300 }
    ]
});
$().w2grid({
    name: 'senderGrid',
    columns: [
        { field: 'email', caption: w2utils.lang('Email'), size: '50%' },
        { field: 'name', caption: w2utils.lang('Name'), size: '50%'}
    ],
    method: 'GET',
    url: '/api/sender',
    onClick: function(event) {
        var grid = this;
        var form = w2ui.senderForm;
        event.onComplete = function () {
            var sel = grid.getSelection();
            if (sel.length == 1) {
                form.recid  = sel[0];
                form.record = $.extend(true, {}, grid.get(sel[0]));
                form.refresh();
            } else {
                form.clear();
            }
        }
    }
});
$().w2form({
    header: w2utils.lang('Edit Record'),
    name: 'senderForm',
    fields: [
        { name: 'recid', type: 'text', html: { caption: 'ID', attr: 'size="10" readonly' } },
        { name: 'name', type: 'text', html: { caption: w2utils.lang('Name'), attr: 'size="40" maxlength="40"' } },
        { name: 'email', type: 'email', required: true, html: { caption: w2utils.lang('Email'), attr: 'size="30"' } }
    ],
    actions: {
        Reset: function () {
            this.clear();
        },
        Save: function () {
            var errors = this.validate();
            if (errors.length > 0) return;
            var cmd;
            var i = this;
            if (i.recid == 0) {
                cmd = 'add'
            } else {
                cmd = 'save'
            }
            console.log(i);
            $.ajax({
                type: "GET",
                url: '/api/sender',
                dataType: "json",
                data: {"request":
                    JSON.stringify({
                        "cmd": cmd,
                        "id": parseInt(i.record.recid),
                        "email": i.record.email,
                        "name": i.record.name
                    }
                )}
            }).done(function (data) {

                if (data['status'] == 'error') {
                    w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                } else {
                    if (i.recid == 0) {
                        i.record.recid = data["recid"];
                        w2ui.senderGrid.add($.extend(true, { recid: data["recid"] }, i.record));
                    } else {
                        w2ui.senderGrid.set(i.recid, i.record);
                    }
                    w2ui.senderGrid.selectNone();
                    i.clear();
                    refreshSenderList($('#campaignSenderId').data('selected').id);
                }
            });

        }
    }
});
// --- /Sender emails editor ---