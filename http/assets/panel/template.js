$('#templateTabs').w2tabs({
    name: 'templateTabs',
    active: 'preview',
    tabs: [
        { id: 'preview', caption: w2utils.lang('Preview') },
        { id: 'code', caption: w2utils.lang('Code') },
        { id: 'help', caption: w2utils.lang('Help') }
    ],
    onClick: function (event) {
        switch (event.target)
        {
            case "preview":
                templateShowPreview();
                break;
            case "code":
                templateShowCode();
                break;
            case "help":
                templateShowHelp();
                break;
        }
    }
});

function templateShowCode() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHelp").hide();
    $("#campaignTemplateCodeContainer").show();
    cm.refresh();
}

function templateShowPreview() {
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelp").hide();
    $('#campaignTemplatePreview').html(cm.getValue());
    $("#campaignTemplatePreviewContainer").show();
}

function templateShowHelp() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelp").show();
}