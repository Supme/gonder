// --- CKEditor ---
var editor = CKEDITOR.replace(
    'campaignTemplate', {
        filebrowserBrowseUrl: '/assets/filemanager/index.html?config=../../filemanager.config',
        extraPlugins: 'codemirror',
        codemirror: {
            autoCloseBrackets: true,
            autoCloseTags: true,
            autoFormatOnStart: true,
            autoFormatOnUncomment: true,
            continueComments: true,
            enableCodeFolding: true,
            enableCodeFormatting: true,
            enableSearchTools: true,
            highlightMatches: true,
            indentWithTabs: false,
            lineNumbers: true,
            lineWrapping: true,
            mode: 'htmlmixed',
            matchBrackets: true,
            matchTags: true,
            showAutoCompleteButton: true,
            showCommentButton: true,
            showFormatButton: true,
            showSearchButton: true,
            showTrailingSpace: true,
            showUncommentButton: true,
            styleActiveLine: true,
            theme: 'default',
            useBeautifyOnStart: false
        },
        plugins:
        'dialog,' +
        'basicstyles,' +
        'clipboard,' +
        'button,' +
        'panelbutton,' +
        'panel,' +
        'floatpanel,' +
        'colorbutton,' +
        'colordialog,' +
        'templates,' +
        'menu,' +
        'contextmenu,' +
        'resize,' +
        'toolbar,' +
        'elementspath,' +
        'enterkey,' +
        'entities,' +
        'popup,' +
        'filebrowser,' +
        'find,' +
        'fakeobjects,' +
        'floatingspace,' +
        'listblock,' +
        'richcombo,' +
        'font,' +
        'format,' +
        'horizontalrule,' +
        'htmlwriter,' +
        'wysiwygarea,' +
        'image,' +
        'indent,' +
        'indentblock,' +
        'indentlist,' +
        'justify,' +
        'menubutton,' +
        'link,' +
        'list,' +
        'liststyle,' +
        'magicline,' +
        'pastetext,' +
        'pastefromword,' +
        'removeformat,' +
        'selectall,' +
        'showblocks,' +
        'showborders,' +
        'sourcearea,' +
        'specialchar,' +
        'stylescombo,' +
        'tab,' +
        'table,' +
        'tabletools,' +
        'undo,' +
        'docprops,',

        allowedContent: true,
        removeFormatAttributes: '',
        height: 450,
        entities: false
    }
);

// --- /CKEditor ---