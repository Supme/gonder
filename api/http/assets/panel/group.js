// --- Group table ---
$('#group').w2grid({
    name: 'group',
    header: w2utils.lang('Group'),
    keyboard : false,
    show: {
        header: true,
        toolbar: true,
        footer: true,
        toolbarSave: true,
        toolbarAdd:true,
        toolbarSearch: false,
    },
    columns: [
        { field: 'recid', caption: w2utils.lang('Id'), size: '50px', sortable: true, style: 'background-color: #efefef; border-bottom: 1px solid white; padding-right: 5px;', attr: "align=right" },
        { field: 'name', caption: w2utils.lang('Name'), size: '100%',  editable: { type: 'text' }}
    ],
    multiSelect: false,
    sortData: [{ field: 'recid', direction: 'DESC' }],
    url: '/api/groups',
    method: 'GET',
    onSelect: function (event) {
        w2ui['campaign'].postData["id"] = parseInt(event.recid);
        w2ui['campaign'].reload();
    },
    onAdd: function (event) {
        var id, name;
        $.ajax({
            type: "GET",
            dataType: 'json',
            data: {"request": JSON.stringify({"cmd": "add"})},
            url: '/api/groups'
        }).done(function(data) {
            if (data['status'] == 'error') {
                w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
            } else {
                id = data["recid"];
                name = data["name"];
                w2ui.group.add({recid: id, name: name},true);
                w2ui.group.editField(id, 1);
            }
        });
    },
    toolbar: {
        items: [
            {id: 'sender', type: 'button', caption: w2utils.lang('Senders'), icon: 'w2ui-icon-pencil'}
        ],
        onClick: function (event) {
            if (event.target == 'sender') {
                if (w2ui['group'].getSelection()[0] == undefined) {
                    w2alert(w2utils.lang('Select group.'));
                } else {
                    w2ui['senderGrid'].postData["id"] = parseInt(w2ui['group'].getSelection()[0]);
                    w2popup.open({
                        title   : w2utils.lang('Sender list editor'),
                        width   : 900,
                        height  : 600,
                        showMax : true,
                        body    : '<div id="senderPopup" style="position: absolute; left: 5px; top: 5px; right: 5px; bottom: 5px;"></div>',
                        onOpen  : function (event) {
                            event.onComplete = function () {
                                $('#w2ui-popup #senderPopup').w2render('senderEditor');
                                w2ui.senderEditor.content('left', w2ui['senderGrid']);
                                w2ui.senderEditor.content('main', w2ui['senderForm']);
                            };
                        },
                        onToggle: function (event) {
                            event.onComplete = function () {
                                w2ui.senderEditor.resize();
                            }
                        }
                    });
                }
            }
        }
    }
});
// --- /Group table ---