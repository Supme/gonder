// --- CKEditor ---
var editor = CKEDITOR.replace(
    'campaignTemplate', {
        filebrowserBrowseUrl: '/assets/filemanager/index.html?config=../../filemanager.config',
        plugins:
        //'dialogui,' +
        'dialog,' +
        'a11yhelp,' +
        'dialogadvtab,' +
        'basicstyles,' +
        'bidi,' +
        'blockquote,' +
        'clipboard,' +
        'button,' +
        'panelbutton,' +
        'panel,' +
        'floatpanel,' +
        'colorbutton,' +
        'colordialog,' +
        'codemirror,' +
        'templates,' +
        'menu,' +
        'contextmenu,' +
        'div,' +
        'resize,' +
        'toolbar,' +
        'elementspath,' +
        'enterkey,' +
        'entities,' +
        'popup,' +
        'filebrowser,' +
        'find,' +
        'fakeobjects,' +
        //'flash,' +
        'floatingspace,' +
        'listblock,' +
        'richcombo,' +
        'font,' +
        //'forms,' +
        'format,' +
        'horizontalrule,' +
        'htmlwriter,' +
        //'iframe,' +
        'wysiwygarea,' +
        'image,' +
        'indent,' +
        'indentblock,' +
        'indentlist,' +
        //'smiley,' +
        'justify,' +
        'menubutton,' +
        //'language,' +
        'link,' +
        'list,' +
        'liststyle,' +
        'magicline,' +
        'maximize,' +
        'pagebreak,' +
        'pastetext,' +
        'pastefromword,' +
        'preview,' +
        'print,' +
        'removeformat,' +
        'selectall,' +
        'showblocks,' +
        'showborders,' +
        'sourcearea,' +
        'specialchar,' +
        'scayt,' +
        'stylescombo,' +
        'tab,' +
        'table,' +
        'tabletools,' +
        'undo,' +
        'wsc,' +
        'docprops,',

        allowedContent:  true,
        removeFormatAttributes: '',
        height: 400,
        entities:false
    }
);

/*
 editor.on( 'instanceReady', function() {
 console.log( editor.filter.allowedContent );
 } );
 */
// --- /CKEditor ---