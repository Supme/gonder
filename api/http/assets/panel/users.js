// --- Users table ---
$().w2grid({
    name: 'userList',
    header: w2utils.lang('Users'),
    show: {
        header: true,
        toolbar: true,
        footer: false,
        toolbarSave: false,
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
    onDblClick: userEditorPopup,
    onAdd: userEditorPopup
});

function userEditorPopup(event){
    //console.log(event);
    var record = w2ui.userList.get(event.recid);

    if (event.type == 'dblClick') {
        w2ui.userEditor.record['id'] = parseInt(event.recid);
        setTimeout(function() {
            $(w2ui.userEditor.get('name').el).prop('disabled', true);
        }, 500);
        w2ui.userEditor.record['name'] = record.name;
    }
    if (event.type == 'add'){}

    w2popup.open({
        width   : 400,
        height  : 480,
        title   : event.type == 'add'?'New user':"Edit user",
        body    : '<div id="userEditor" style="width: 100%; height: 100%;"></div>',
        onOpen  : function (event) {
            event.onComplete = function () {
                $('#userEditor').w2render('userEditor');
            }
        },
        onClose : function (event) {
            w2ui.userEditor.clear();
        }
    });

    var unit = [];
    $.each(w2ui.unitList.records, function(k, v){
        unit[k] = {id:v.recid, text:v.name}
    });
    w2ui.userEditor.set('unit', {options: {items: unit}});
    if (event.type == 'dblClick') {
        setTimeout(function () {
            $('#userEditor #unit').w2field().setIndex(findRecId(unit, record.unitid));
        }, 500);
    }

    var groups = [];
    $.ajax({
        type: "GET",
        dataType: 'json',
        data: {"request": JSON.stringify({"cmd": "get"})},
        url: '/api/groups'
    }).done(function(data) {
        $.each(data.records, function(k, v){
            groups[k] = {id:v.recid, text:v.name}
        });
    });
    w2ui.userEditor.set('group', {options: {items: groups, openOnFocus:true}});
    if (event.type == 'dblClick') {
        setTimeout(function() {
            $.each(record.groupsid, function (k, v) {
                $('#userEditor #group').w2field().setIndex(findRecId(groups, v), true);
            })
        }, 500);
    }
    w2ui.userEditor.refresh();
}

function findRecId(data, id) {
    var i = false;
    $.each(data, function (k, v) {
        if(v.id == id) {
            i = k;
        }
    });
    return i;
}

$().w2form({
    name: 'userEditor',
    fields: [
        {name: 'name', html: {caption: 'Name'}, type: 'text'},
        {name: 'password', html: {caption: w2utils.lang('Password')}, type: 'pass'},
        {name: 'unit', html: {caption: w2utils.lang('Unit')}, type: 'list'},
        {name: 'group', html: {caption: w2utils.lang('Group')}, type: 'enum'}//, required: true}
    ],
    url: 'api/users',
    method: 'POST',
    onError: function(event){
        // ToDo alert not close...
        w2alert(w2utils.lang(event.message));
    },
    actions: {
        save: function () {
            this.save(this.record.id == undefined?{cmd:'add'}:undefined, function () {
                w2ui['userList'].reload();
                w2popup.close();
            });
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
