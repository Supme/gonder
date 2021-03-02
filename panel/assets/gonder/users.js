// --- Users table ---
$().w2grid({
    name: 'userList',
    header: w2utils.lang('Users'),
    show: {
        header: true,
        toolbar: true,
        footer: true,
        toolbarSave: false,
        toolbarAdd: true,
        toolbarSearch: false
    },
    columns: [
        {field: 'name', text: w2utils.lang('Name'), size: '50%', sortable: true}
    ],
    multiSelect: false,
    sortData: [{field: 'recid', direction: 'ASC'}],
    method: 'GET',
    postData: { cmd: "get" },
    onDblClick: userEditorPopup,
    onAdd: userEditorPopup
});

function userEditorPopup(event){
    var record = w2ui.userList.get(event.recid);

    if (event.type === 'dblClick') {
        w2ui.userEditor.record['id'] = parseInt(event.recid);
        setTimeout(function() {
            $(w2ui.userEditor.get('name').el).prop('disabled', true);
        }, 500);
        w2ui.userEditor.record['name'] = record.name;
    }
    if (event.type === 'add'){}

    w2popup.open({
        name: "userEditor",
        width   : 400,
        height  : 480,
        title   : event.type === 'add'?'New user':"Edit user",
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
        unit[k] = {id: v.recid, text: v.name}
    });
    w2ui['userEditor'].set('unit', { options: { items: unit } });
    w2ui['userEditor'].record['unit'] = record.unitid;

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
        w2ui.userEditor.refresh();
    });
    w2ui.userEditor.set('group', { options: { items: groups, openOnFocus:true } });
    if (event.type === 'dblClick') {
        setTimeout(function() {
            $.each(record.groupsid, function (k, v) {
                $('#userEditor #group').w2field().setIndex(findKey(groups, v), true);
            })
        }, 100);
    }
}

function findKey(data, id) {
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
        { field: 'name', html: { label: 'Name' }, type: 'text' },
        { field: 'password', html: { label: 'Password'}, type: 'pass' },
        { field: 'unit', html: { label: 'Unit' }, type: 'list'},
        { field: 'group', html: { label: 'Group' }, type: 'enum' }//, required: true}
    ],
    url: 'api/users',
    method: 'POST',
    onError: function(event){
        // ToDo alert not close...
        w2alert(w2utils.lang(event.message));
    },
    actions: {
        save: function () {
            if (this.record.group == undefined) this.record.group = [];
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
        footer: true,
        toolbar: true,
        toolbarAdd: true
    },
    columns: [
        {field: 'name', text: w2utils.lang('Name'), size: '100%', sortable: true }
    ],
    multiSelect: false,
    sortData: [{field: 'recid', direction: 'ASC'}],
    method: 'GET',
    postData: { cmd:"get" },
    onDblClick: unitEditorPopup,
    onAdd: unitEditorPopup
});

$().w2grid({
    name: 'groupList',
    header: w2utils.lang('Groups'),
    show: {
        header: true,
        selectColumn: true
    },
    columns: [
        { field: 'recid', text: w2utils.lang('Id'), size: '50px', sortable: true },
        { field: 'name', text: w2utils.lang('Name'), size: '100%' }
    ],
    sortData: [{field: 'recid', direction: 'ASC'}],
    url: '/api/groups',
    method: 'GET',
    postData: { cmd:"get" }
});

function unitEditorPopup(event) {
    var record = w2ui.unitList.get(event.recid);

    if (event.type === 'dblClick') {
        w2ui.unitEditor.record['id'] = parseInt(event.recid);
        w2ui.unitEditor.record['name'] = record.name;
    }
    if (event.type === 'add') {
    }

    w2popup.open({
        width: 600,
        height: 350,
        title: event.type === 'add' ? 'New unit' : "Edit unit",
        body: '<div id="unitEditor" style="width: 100%; height: 100%;"></div>',
        onOpen: function (event) {
            event.onComplete = function () {
                $('#unitEditor').w2render('unitEditor');
            }
        },
        onClose: function (event) {
            w2ui.unitEditor.clear();
        }
    });

    if (event.type === 'dblClick') {
        $.ajax({
            type: "GET",
            dataType: 'json',
            data: {"request": JSON.stringify({"cmd": "rights", "id": parseInt(event.recid)})},
            url: '/api/units'
        }).done(function(data) {
            $.each(data, function(k, v){
                w2ui.unitEditor.record[k] = v;
            });
            w2ui.unitEditor.refresh();
        });
    }
}

$().w2form({
    name: 'unitEditor',
    tabs: [
        { id: 'tab1', text: w2utils.lang('General') },
        { id: 'tab2', text: w2utils.lang('Group right') },
        { id: 'tab3', text: w2utils.lang('Campaign right') },
        { id: 'tab4', text: w2utils.lang('Recipients right') },
        { id: 'tab5', text: w2utils.lang('Profile right') }
    ],
    fields: [
        {field: 'name', html: { label: 'Name' }, type: 'text' },

        {field: 'get-groups', html: { label: 'Get groups', page: 1, column: 0 }, type: 'checkbox' },
        {field: 'save-groups', html: { label: 'Save groups', page: 1, column: 0 }, type: 'checkbox' },
        {field: 'add-groups', html: { label: 'Add groups', page: 1, column: 0 }, type: 'checkbox' },
        {field: 'save-campaigns', html: { label: 'Save campaigns', page: 2, column: 0 }, type: 'checkbox' },
        {field: 'add-campaigns', html: { label: 'Add campaigns', page: 2, column: 0 }, type: 'checkbox' },
        {field: 'get-campaigns', html: { label:'Get campaigns', page: 2, column: 0 }, type: 'checkbox' },
        {field: 'get-campaign', html: { label: 'Get campaign', page: 2, column: 1 }, type: 'checkbox' },
        {field: 'save-campaign', html: { label: 'Save campaign', page: 2, column: 1 }, type: 'checkbox' },
        {field: 'get-recipients', html: { label: 'Get recipients', page: 3, column: 0 }, type: 'checkbox' },
        {field: 'get-recipient-parameters', html: { label: 'Get recipient params', page: 3, column: 0 }, type: 'checkbox' },
        {field: 'upload-recipients', html: { label: 'Upload recipients', page: 3, column: 1 }, type: 'checkbox' },
        {field: 'delete-recipients', html: { label: 'Delete recipients', page: 3, column: 1 }, type: 'checkbox' },
        {field: 'get-profiles', html: { label: 'Get profiles', page: 4, column: 0 }, type: 'checkbox' },
        {field: 'add-profiles', html: { label: 'Add profiles', page: 4, column: 0 }, type: 'checkbox' },
        {field: 'delete-profiles', html: { label: 'Delete profiles', page: 4, column: 0 }, type: 'checkbox' },
        {field: 'save-profiles', html: { label: 'Save profiles', page: 4, column: 0 }, type: 'checkbox' },
        {field: 'accept-campaign', html: { label: 'Accept campaign', page: 2, column: 1 }, type: 'checkbox' },
        {field: 'get-log-main', html: { label: 'Get log main', page: 0, column: 0 }, type: 'checkbox' },
        {field: 'get-log-api', html: { label: 'Get log api', page: 0, column: 0 }, type: 'checkbox' },
        {field: 'get-log-campaign', html: { label: 'Get log campaign', page: 0, column: 0 }, type: 'checkbox' },
        {field: 'get-log-utm', html: { label: 'Get log utm', page: 0, column: 0 }, type: 'checkbox' }
    ],
    url: 'api/units',
    method: 'POST',
    onError: function(event){
        // ToDo alert not close...
        w2alert(w2utils.lang(event.message));
    },
    actions: {
        save: function () {
            this.save(this.record.id == undefined?{cmd:'add'}:undefined, function () {
                w2ui['unitList'].reload();
                w2popup.close();
            });
        }
    }
});




$().w2layout({
    name: 'users',
    panels: [
        { type: 'main', resizable: true, size: "50%" },
        { type: 'right', resizable: true, size: "50%" }
    ]
});

$("#users").w2render('users');
w2ui.users.html('main', w2ui.userList);
w2ui.users.html('right', w2ui.unitList);
