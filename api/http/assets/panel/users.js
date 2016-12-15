// --- Users table ---
$().w2grid({
    name: 'userList',
    header: w2utils.lang('Users'),
    show: {
        header: true,
        toolbar: true,
        footer: false,
        toolbarSave: true,
        toolbarAdd: true,
        toolbarSearch: false
    },
    columns: [
        {field: 'name', caption: w2utils.lang('Name'), size: '50%', editable: {type: 'text'}}
    ],
    multiSelect: false,
    sortData: [{field: 'recid', direction: 'ASC'}],
    url: '/api/users',
    method: 'GET',
    onDblClick: function (event){
        w2popup.open({
            width   : 400,
            height  : 480,
            title   : w2utils.lang('User editor'),
            body    : '<div id="userEditor" style="width: 100%; height: 100%;"></div>',
            onOpen  : function (event) {
                event.onComplete = function () {
                    $('#userEditor').w2render('userEditor');
                }
            }
        });
        w2ui.userEditor.clear();

        var record = w2ui.userList.get(event.recid);
        w2ui.userEditor.record['userEditorId'] = event.recid;
        w2ui.userEditor.record['userEditorName'] = record.name;
        w2ui.userEditor.record['userEditorPassword'] = record.password;

        var unit = [];
        $.each(w2ui.unitList.records, function(k, v){
            unit[k] = {id:v.recid, text:v.name}
        });
        w2ui.userEditor.set('userEditorUnit', {options: {items: unit}});
        setTimeout(function () {
            $('#userEditorUnit').w2field().setIndex(findRecId(unit, record.unitid));
        }, 500);

        var groups = [];
        $.ajax({
            type: "GET",
            dataType: 'json',
            data: {"request": JSON.stringify({"cmd": "get"})},
            url: '/api/groups'
        }).done(function(data) {
            console.log(data);
            $.each(data.records, function(k, v){
                groups[k] = {id:v.recid, text:v.name}
            });
        });
        console.log(groups);
        w2ui.userEditor.set('userEditorGroup', {options: {items: groups, openOnFocus:true}});
        setTimeout(function() {
            $.each(record.groupsid, function (k, v) {
                $('#userEditorGroup').w2field().setIndex(findRecId(groups, v), true);
            })
        }, 500);
        w2ui.userEditor.refresh();
    },
    onSave: function (event) {
        console.log(event);
        w2ui.userList.reload();
    }
});

function findRecId(data, id) {
    var i = false;
    $.each(data, function (k, v) {
        if(v.id == id) {
            console.log('find id '+id+' in '+k+' element');
            i = k;
        }
    });
    return i;
}

$().w2form({
    name: 'userEditor',
    fields: [
        {name: 'userEditorName', html: {caption: w2utils.lang('Name'), attr: 'readonly'}, type: 'text'},
        {name: 'userEditorPassword', caption: w2utils.lang('Password'), type: 'pass'},
        {name: 'userEditorUnit', caption: w2utils.lang('Unit'), type: 'list'},
        {name: 'userEditorGroup', caption: w2utils.lang('Group'), type: 'enum'}
    ],
    url: 'api/users',
    method: 'POST',
    actions: {
        reset: function () {
            this.clear();
        },
        save: function () {
            this.save();
        }
    }
});

$().w2grid({
    name: 'unitList',
    header: w2utils.lang('Units'),
    show: {
        header: true,
        toolbar: true,
//        selectColumn: true,
        toolbarSave: true,
        toolbarAdd: true
    },
    columns: [
        {field: 'name', caption: w2utils.lang('Name'), size: '100%'}
    ],
    multiSelect: false,
    sortData: [{field: 'recid', direction: 'ASC'}],
    url: '/api/units',
    method: 'GET'
});

$().w2grid({
    name: 'groupList',
    header: w2utils.lang('Groups'),
    show: {
        header: true,
        selectColumn: true
    },
    columns: [
        { field: 'recid', caption: w2utils.lang('Id'), size: '50px', sortable: true },
        { field: 'name', caption: w2utils.lang('Name'), size: '100%' }
    ],
    sortData: [{field: 'recid', direction: 'ASC'}],
    url: '/api/groups',
    method: 'GET'
});

$().w2layout({
    name: 'users',
    panels: [
        { type: 'main', resizable: true, size: "50%" },
        { type: 'right', resizable: true, size: "50%" }
    ]
});

$("#users").w2render('users');
w2ui.users.content('main', w2ui.userList);
w2ui.users.content('right', w2ui.unitList);
