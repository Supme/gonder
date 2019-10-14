function loadLink(url) {
    var link=document.createElement('a');
    document.body.appendChild(link);
    link.href=url;
    link.click();
}