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
        {field: 'name', caption: w2utils.lang('Name'), size: '50%', sortable: true}
    ],
    multiSelect: false,
    sortData: [{field: 'recid', direction: 'ASC'}],
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
        {field: 'name', caption: w2utils.lang('Name'), size: '100%', sortable: true }
    ],
    multiSelect: false,
    sortData: [{field: 'recid', direction: 'ASC'}],
    method: 'GET',
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
        { field: 'recid', caption: w2utils.lang('Id'), size: '50px', sortable: true },
        { field: 'name', caption: w2utils.lang('Name'), size: '100%' }
    ],
    sortData: [{field: 'recid', direction: 'ASC'}],
    url: '/api/groups',
    method: 'GET'
});

function unitEditorPopup(event) {
    //console.log(event);
    var record = w2ui.unitList.get(event.recid);

    if (event.type == 'dblClick') {
        w2ui.unitEditor.record['id'] = parseInt(event.recid);
        w2ui.unitEditor.record['name'] = record.name;
    }
    if (event.type == 'add') {
    }

    w2popup.open({
        width: 600,
        height: 350,
        title: event.type == 'add' ? 'New unit' : "Edit unit",
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

    if (event.type == 'dblClick') {
        $.ajax({
            type: "GET",
            dataType: 'json',
            data: {"request": JSON.stringify({"cmd": "rights", "id": parseInt(event.recid)})},
            url: '/api/units'
        }).done(function(data) {
            $.each(data, function(k, v){
                w2ui.unitEditor.record[k] = v;
            });
        });
    }
    w2ui.unitEditor.refresh();
}

$().w2form({
    name: 'unitEditor',
    tabs: [
        { id: 'tab1', caption: w2utils.lang('General') },
        { id: 'tab2', caption: w2utils.lang('Group right') },
        { id: 'tab3', caption: w2utils.lang('Campaign right') },
        { id: 'tab4', caption: w2utils.lang('Recipients right') },
        { id: 'tab5', caption: w2utils.lang('Profile right') }
    ],
    fields: [
        {name: 'name', html: {caption: 'Name'}, type: 'text'},

        {name: 'get-groups', html: {caption: w2utils.lang('Get groups'), page: 1, column: 0 }, type: 'checkbox'},
        {name: 'save-groups', html: {caption: w2utils.lang('Save groups'), page: 1, column: 0}, type: 'checkbox'},
        {name: 'add-groups', html: {caption: w2utils.lang('Add groups'), page: 1, column: 0}, type: 'checkbox'},
        {name: 'save-campaigns', html: {caption: w2utils.lang('Save campaigns'), page: 2, column: 0}, type: 'checkbox'},
        {name: 'add-campaigns', html: {caption: w2utils.lang('Add campaigns'), page: 2, column: 0}, type: 'checkbox'},
        {name: 'get-campaigns', html: {caption: w2utils.lang('Get campaigns'), page: 2, column: 0}, type: 'checkbox'},
        {name: 'get-campaign', html: {caption: w2utils.lang('Get campaign'), page: 2, column: 1}, type: 'checkbox'},
        {name: 'save-campaign', html: {caption: w2utils.lang('Save campaign'), page: 2, column: 1}, type: 'checkbox'},
        {name: 'get-recipients', html: {caption: w2utils.lang('Get recipients'), page: 3, column: 0}, type: 'checkbox'},
        {name: 'get-recipient-parameters', html: {caption: w2utils.lang('Get recipient parameters'), page: 3, column: 0}, type: 'checkbox'},
        {name: 'upload-recipients', html: {caption: w2utils.lang('Upload recipients'), page: 3, column: 1}, type: 'checkbox'},
        {name: 'delete-recipients', html: {caption: w2utils.lang('Delete recipients'), page: 3, column: 1}, type: 'checkbox'},
        {name: 'get-profiles', html: {caption: w2utils.lang('Get profiles'), page: 4, column: 0}, type: 'checkbox'},
        {name: 'add-profiles', html: {caption: w2utils.lang('Add profiles'), page: 4, column: 0}, type: 'checkbox'},
        {name: 'delete-profiles', html: {caption: w2utils.lang('Delete profiles'), page: 4, column: 0}, type: 'checkbox'},
        {name: 'save-profiles', html: {caption: w2utils.lang('Save profiles'), page: 4, column: 0}, type: 'checkbox'},
        {name: 'accept-campaign', html: {caption: w2utils.lang('Accept campaign'), page: 2, column: 1}, type: 'checkbox'},
        {name: 'get-log-main', html: {caption: w2utils.lang('Get log main'), page: 0, column: 0}, type: 'checkbox'},
        {name: 'get-log-api', html: {caption: w2utils.lang('Get log api'), page: 0, column: 0}, type: 'checkbox'},
        {name: 'get-log-campaign', html: {caption: w2utils.lang('Get log campaign'), page: 0, column: 0}, type: 'checkbox'},
        {name: 'get-log-utm', html: {caption: w2utils.lang('Get log utm'), page: 0, column: 0}, type: 'checkbox'}
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
w2ui.users.content('main', w2ui.userList);
w2ui.users.content('right', w2ui.unitList);
