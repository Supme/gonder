var cm = CodeMirror.fromTextArea(document.getElementById("campaignTemplateCode"), {
    lineNumbers: true,
    mode: {
        name: "htmlmixed",
        scriptTypes: [{matches: /\/x-handlebars-template|\/x-mustache/i,mode: null}]
    },
    selectionPointer: true,
    theme: "dracula"
});

var gjs = grapesjs.init({
    container : '#campaignTemplateGrapesjs',
    storageManager: { type: null },
    plugins: ['gjs-preset-newsletter'],
    pluginsOpts: {
        'gjs-preset-newsletter': {
            //modalTitleImport: 'Import template',
        }
    },
    assetManager: {
        autoAdd: 1,
        upload: "./api/files",
        params: {"testparam": "testvalue"},
        uploadName: 'files'
    }
});

var am = gjs.AssetManager;

// The upload is started
gjs.on('asset:upload:start', function() {
    // am.setParameter("./api/", "nm", "vl");
    console.log("Upload start");
});

// The upload is ended (completed or not)
gjs.on('asset:upload:end', function() {
    console.log("Upload end");
});

// Error handling
gjs.on('asset:upload:error', function(err) {
    console.log("Upload error: " + err);
});

// Do something on response
gjs.on('asset:upload:response', function(response) {
    console.log("Upload response: " + response);
});

gjs.on('asset:remove', function(f) {
    console.log("Remove: " + f.attributes["path"]);
    console.log(f);
    // ToDo request to backend for remove file
});


gjs.on('run:open-assets', function() {
    console.log("Run open assets");
    am.getAll().reset();
    // ToDo request to backend for files list
    am.add([
        {
            // You can pass any custom property you want
            path: '/350x250/78c5d6/fff/image1.jpg',
            src: 'http://placehold.it/350x250/78c5d6/fff/image1.jpg',
        }, {
            path: '/350x250/459ba8/fff/image2.jpg',
            src: 'http://placehold.it/350x250/459ba8/fff/image2.jpg',
        }, {
            path: '/350x250/79c267/fff/image3.jpg',
            src: 'http://placehold.it/350x250/79c267/fff/image3.jpg',
        }
        // ...
    ]);
    am.render();
});

$("#campaignTemplateGrapesjsButton").html(
    '<button onclick="templateGrapesjsExport()">' + w2utils.lang("Export to code") + '</button>&nbsp;' +
    '<button onclick="templateGrapesjsImport();">' + w2utils.lang("Import from code") + '</button>&nbsp;' +
    '<button onclick="templateGrapesjsClear();">' + w2utils.lang("Clear") + '</button>&nbsp;'
);

$('#templateTabs').w2tabs({
    name: 'templateTabs',
    active: 'preview',
    tabs: [
        { id: 'preview', caption: w2utils.lang('Preview') },
        { id: 'code', caption: w2utils.lang('Code') },
        { id: 'help', caption: w2utils.lang('Help') },
//        { id: 'grapesjs', caption: w2utils.lang('GrapesJs') }
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
            case "grapesjs":
                templateShowGrapesjs();
                break;
        }
    }
});

function templateShowCode() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateGrapesjsContainer").hide();
    $("#campaignTemplateCodeContainer").show();
    cm.refresh();
}

function templateShowPreview() {
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateGrapesjsContainer").hide();
    $('#campaignTemplatePreview').html(cm.getValue());
    $("#campaignTemplatePreviewContainer").show();
}

function templateShowHelp() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateGrapesjsContainer").hide();
    $("#campaignTemplateHelpContainer").show();
}

function templateShowGrapesjs() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateGrapesjsContainer").show();



}

function templateGrapesjsImport() {
    w2confirm(w2utils.lang('Import template from code to GrapesJs?'), function (btn) {
        if (btn == 'Yes') {
            var el = document.createElement('html');
            el.innerHTML = cm.getValue();
            var body = el.getElementsByTagName('body');
            gjs.setComponents(body[0].innerHTML);
        }
    });
}

function templateGrapesjsExport() {
    w2confirm(w2utils.lang('Export GrapesJs template to code?'), function (btn) {
        if (btn == 'Yes') {
            var code = "" +
                "<!DOCTYPE html>\n" +
                "<html>\n" +
                "  <head>\n" +
                "    <meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" />\n" +
                "    <meta name=\"viewport\" content=\"width=device-width; initial-scale=1.0; maximum-scale=1.0;\" />\n" +
                "    <title>" + $("#campaignSubject").val() + "</title>\n" +
                "    <style>\n" +
                "      " + gjs.getCss() + "\n" +
                "    </style>" +
                "  </head>\n" +
                "  <body bgcolor=\"#ffffff\" style=\"margin: 0; padding: 0;\">\n" +
                "    " + gjs.getHtml() + "\n" +
                "  </body>\n" +
                "</html>\n";
            cm.setValue(code);
        }
    });

}

function templateGrapesjsClear() {
    w2confirm(w2utils.lang('Clear GrapesJs template?'), function (btn) {
        if (btn == 'Yes') {
            gjs.setComponents("");
        }
    });
}
// cm.getValue()
// gjs.getHtml()
// cm.setValue(gjs.getHtml())
//