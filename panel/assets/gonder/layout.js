// --- Layout ---
$('#layout').w2layout({
    name: 'layout',
    panels: [
        { type: 'top', size: 35, style: "padding: 4px; border: 1px solid #dfdfdf; border-radius: 3px;" },
        { type: 'main', hidden: true},
        { type: 'bottom', size: 250, resizable: true }
    ]
});

w2ui.layout.content('top', $().w2toolbar({
    name: 'toolbar',
    right : 'v' + version,
    items: [
        { type: 'radio', id: 'parametersButton', group: '1', text: w2utils.lang('Parameters'), img: 'icon-page' },
        { type: 'radio', id: 'editorButton', group: '1', text: w2utils.lang('Editor'), img: 'w2ui-icon-pencil' },
        { type: 'radio', id: 'recipientsButton', group: '1', text: w2utils.lang('Recipients'), img: 'w2ui-icon-columns' },
        { type: 'break' },
        { type: 'check', id: 'acceptSend', group: '1', text: w2utils.lang('Accept send'), style: '.checked {background: #ddff00}' },
        { type: 'spacer' },
        { type: 'radio', id: 'statusButton', group: '1', text: w2utils.lang('Status'), img: 'w2ui-icon-info' },
        { type: 'radio', id: 'usersButton', group: '1', text: w2utils.lang('Users'), img: 'w2ui-icon-columns' },
        { type: 'break' }
    ],
    onClick: function (event) {
        // console.log('Target: '+ event.target, event);
        if (
            $('#campaignId').val() !== '' ||
            event.target ==='statusButton' ||
            event.target ==='usersButton'
        ) {
            switch (event.target) {
                case 'parametersButton':
                    switchToParameters();
                    break;
                case 'editorButton':
                    switchToEditor();
                    break;
                case 'recipientsButton':
                    switchToRecipients();
                    break;

                case 'saveButton':
                    saveCampaign();
                    break;

                case 'acceptSend':
                    switchAcceptSend();
                    break;

                case 'usersButton':
                    switchToUsers();
                    break;
                case 'statusButton':
                    switchToStatus();
                    break;

            }
            w2ui['layout'].resize();
        } else {
            console.log(event.target);
            event.checked = false;
            w2alert(w2utils.lang('Select group and campaign, before select action.'));
        }
    }
}));

w2ui.layout.content('main', $('#mainbox').html());
w2ui.layout.content('bottom', $().w2layout({
    name: 'bottom',
    panels: [
        { type: 'left', size: '50%', resizable: true },
        { type: 'main',  size: '50%', resizable: true }
    ]
}));

function switchAcceptSend() {
    var campaignData = getCampaignData($('#campaignId').val());
    var confirm;
    if (w2ui['toolbar'].get('acceptSend').checked) {
        confirm = 'Are you sure to deactivate campaign?';
    } else {
        confirm = 'Are you sure to activate campaign?';
    }
    w2popup.open({
        title: w2utils.lang(confirm),
        body:
            '<div>' +
            '<div>' + w2utils.lang("Subject") + ': "<b>' + campaignData.subject + '</b>"</div>' +
            '<div>' + w2utils.lang("Sender") + ': "<b>' + campaignData.senderName + '</b>"</div>' +
            '<div>' + w2utils.lang("Profile") + ': "' + campaignData.profileName + '"</div>' +
            '<div>' + w2utils.lang("Start date") + ': "' + w2utils.formatDate(campaignData.startDate, w2utils.settings.dateFormat) + ' ' + w2utils.formatTime(campaignData.startDate, w2utils.settings.timeFormat) +'"</div>' +
            '<div>' + w2utils.lang("End date") + ': "' + w2utils.formatDate(campaignData.endDate, w2utils.settings.dateFormat) + ' ' + w2utils.formatTime(campaignData.endDate, w2utils.settings.timeFormat) +'"</div>' +
            '<br>' +
            '<div>' +
            ' <div style="float: left; width: 420px">' +
            '  <div style="position: absolute; border: 1px solid #333; width: 420px; height: 300px; overflow-y: scroll;">' + campaignData.templateHTML + '</div>' +
            ' </div>' +
            ' <div style="float: right;width: 360px;">' +
            '  <div style="position: absolute; border: 1px solid #333; width: 360px; height: 300px; overflow-y: scroll; background-color: #fff"><pre>' + campaignData.templateText + '</pre></div>' +
            ' </div>' +
            '</div>'
        ,
        buttons: '<button class="w2ui-btn" onclick="changeAcceptSend(true); w2popup.close();">'+ w2utils.lang("Yes") + '</button>'+
                 '&nbsp;'+
                 '<button class="w2ui-btn" onclick="changeAcceptSend(false); w2popup.close();">'+ w2utils.lang("No") + '</button>',
        width: 800,
        height: 480,
        showMax: false,
        showClose: false,
        keyboard: false
    })
}

function changeAcceptSend(change) {
    var accept = w2ui['toolbar'].get('acceptSend').checked;
    if (change) {
        $.ajax({
            type: "POST",
            url: '/api/campaign',
            dataType: "json",
            data: {"request": JSON.stringify({"cmd": "accept", "id": parseInt($('#campaignId').val()), "select": w2ui['toolbar'].get('acceptSend').checked})}
        }).done(function (data) {
            if (data['status'] === 'error') {
                accept = !accept;
                w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
            }
        })
    } else {
        accept = !accept;
    }
    setAcceptSend(accept);
}

function setAcceptSend(accept) {
    if (accept) {
        w2ui['toolbar'].check('acceptSend');
        w2ui['toolbar'].get('acceptSend').text = w2utils.lang('Cancel accepted send');
    } else {
        w2ui['toolbar'].uncheck('acceptSend');
        w2ui['toolbar'].get('acceptSend').text = w2utils.lang('Accept send');
    }
    w2ui['toolbar'].refresh('acceptSend');
}

function switchToParameters() {
    $('#parameterTemplatePreview').html(cmHTML.getValue())
    $('#template').hide();
    $('#recipient').hide();
    $('#profile').hide();
    $('#users').hide();
    $('#status').hide();
    $('#parameter').show();
}

function switchToEditor() {
    $('#parameter').hide();
    $('#recipient').hide();
    $('#template').show();
    w2ui.templateTabs.click('preview');
    $('#profile').hide();
    $('#users').hide();
    $('#status').hide();
}

function switchToRecipients() {
    w2ui['recipient'].url = '/api/recipients';
    w2ui['recipient'].reload();
    $('#template').hide();
    $('#parameter').hide();
    $('#recipient').show();
    $('#profile').hide();
    $('#users').hide();
    $('#status').hide();
}

function switchToUsers() {
    w2ui['userList'].url = '/api/users';
    w2ui['userList'].reload();
    w2ui['unitList'].url = '/api/units';
    w2ui['unitList'].reload();
    $('#template').hide();
    $('#parameter').hide();
    $('#recipient').hide();
    $('#profile').hide();
    $('#users').show();
    $('#status').hide();
}

function switchToStatus() {
    $('#template').hide();
    $('#parameter').hide();
    $('#recipient').hide();
    $('#profile').hide();
    $('#users').hide();
    $('#status').show();
}
// --- /Layout ---

// --- Parameters form ---
$('#parameterForm').w2form({
    name: 'parameter',
    fields: [
        { name: 'campaignId', type: 'text', html: { caption: w2utils.lang('Id'), attr: 'size="4" readonly' } },
        { name: 'campaignName', type: 'text', html: { caption: w2utils.lang('Name'), attr: 'size="40" readonly' } },
        { name: 'campaignProfileId', type: 'list', html: { caption: w2utils.lang('Profile'), attr: 'size="40"' }, minLength: 0},
        { name: 'campaignSubject', type: 'text', html: { caption: w2utils.lang('Subject'), attr: 'size="40" autocomplete="off"' } },
        { name: 'campaignSenderId', type: 'list', html: { caption: w2utils.lang('Sender'), attr: 'size="40"' }, minLength: 0},
        { name: 'campaignStartDate', type: 'date', html: { caption: w2utils.lang('Start date'), attr: 'size="10" autocomplete="off"' }, options: {format: w2utils.settings.dateFormat} },
        { name: 'campaignStartTime', type: 'time', html: { caption: w2utils.lang('Start time'), attr: 'size="10" autocomplete="off"'}, options: {format: w2utils.settings.timeFormat} },
        { name: 'campaignEndDate', type: 'date', html: { caption: w2utils.lang('End date'), attr: 'size="10" autocomplete="off"' }, options: {format: w2utils.settings.dateFormat} },
        { name: 'campaignEndTime', type: 'time', html: { caption: w2utils.lang('End time'), attr: 'size="10" autocomplete="off"' }, options: {format: w2utils.settings.timeFormat} },
        { name: 'campaignCompressHTML', type: 'checkbox', html: { caption: w2utils.lang('Compress HTML') } },
        { name: 'campaignSendUnsubscribe', type: 'checkbox', html: { caption: w2utils.lang('Send unsubscribe') } },
    ],
    actions: {
        Save: function () {
            saveCampaign();
        }
    }
});

$('#campaignSendUnsubscribe').click(function(data) {
    if ($('#campaignSendUnsubscribe').is(':checked')) {
        w2confirm(w2utils.lang('Are you sure send mail for unsubscribed?'), function (btn) {
            if (btn !== 'Yes') {
                $('#campaignSendUnsubscribe').prop('checked', !$('#campaignSendUnsubscribe').is(':checked'));
            }
        });
    }
});
// --- /Parameters form ---

