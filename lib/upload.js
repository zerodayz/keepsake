document.querySelector('#file-upload-field').addEventListener('change', function(){
    if (this.files.length === 0)
        return;

    const name = document.querySelector('#file-upload-field').files[0].name;

    document.getElementsByClassName("file-upload-wrapper")[0].setAttribute("data-text", name);
}, false);