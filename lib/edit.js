var modal = document.getElementById("Modal");
var modalCreate = document.getElementById("ModalCreate");
var modalEdit = document.getElementById("ModalEdit");
var modalHistory = document.getElementById("ModalHistory");

const classA = Array.from(document.getElementsByClassName("search-collapsible"))
    ,classB = Array.from(document.getElementsByClassName("search-no-collapsible"))
    ,SearchTitles = Array.from(new Set(classA.concat(classB)))
var i;

for (i = 0; i < SearchTitles.length; i++) {
    SearchTitles[i].addEventListener("click", function () {
        this.classList.toggle("search-active");
        var content = this.nextElementSibling;
        if (content.style.display === "block") {
            content.style.display = "none";
        } else {
            content.style.display = "block";
        }
    });
}

if (modal) {
    var btn = document.getElementById("ModalButton");
    var span = document.getElementsByClassName("close")[0];
    if (btn) {
        btn.onclick = function () {
            modal.style.display = "block";
        }
        span.onclick = function () {
            modal.style.display = "none";
        }
    }
}
if (modalCreate) {
    var btnCreate = document.getElementById("ModalButtonCreate");
    var spanCreate = document.getElementsByClassName("close-create")[0];
    if (btnCreate) {
        btnCreate.onclick = function () {
            modalCreate.style.display = "block";
        }
        spanCreate.onclick = function () {
            modalCreate.style.display = "none";
        }
    }
}
if (modalEdit) {
    var btnEdit = document.getElementById("ModalButtonEdit");
    var spanEdit = document.getElementsByClassName("close-edit")[0];
    if (btnEdit) {
        btnEdit.onclick = function () {
            modalEdit.style.display = "block";
        }
        spanEdit.onclick = function () {
            modalEdit.style.display = "none";
        }
    }
}
if (modalHistory) {
    var btnHistory = document.getElementById("ModalButtonHistory");
    var spanHistory = document.getElementsByClassName("close-history")[0];
    if (btnHistory) {
        btnHistory.onclick = function () {
            modalHistory.style.display = "block";
        }
        spanHistory.onclick = function () {
            modalHistory.style.display = "none";
        }
    }
}

window.onclick = function (event) {
    if (event.target == modal) {
        modal.style.display = "none";
    }
    if (event.target == modalCreate) {
        modalCreate.style.display = "none";
    }
    if (event.target == modalEdit) {
        modalEdit.style.display = "none";
    }
    if (event.target == modalHistory) {
        modalHistory.style.display = "none";
    }
}

function validate_form() {
    var i = 0;
    i = 0;

    var button = document.getElementById('submit-button');

    var comment_title = document.getElementById('comment_title_value').value
    var RegExp_comment_title = /^[a-zA-Z0-9_\[\]!@#$%^&*()\-+=\\'";:<,>./?]/;
    var validTitle = RegExp_comment_title.test(comment_title);

    if (validTitle) {
        document.getElementById('comment_title').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightgreen\'>Valid Title</span>';
        i += 1
    } else {
        document.getElementById('comment_title').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightcoral\'>Invalid Title</span>';
    }

    var comment_message = document.getElementById('comment_message_value').value
    var RegExp_comment_message = /^[a-zA-Z0-9_\[\]!@#$%^&*()\-+=\\'";:<,>./?]/;
    var validMessage = RegExp_comment_message.test(comment_message);

    if (validMessage) {
        document.getElementById('comment_message').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightgreen\'>Valid Message</span>';
        i += 1
    } else {
        document.getElementById('comment_message').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightcoral\'>Invalid Message</span>';
    }

    if (i == 2) {
        button.disabled = false;
    } else {
        button.disabled = true;
    }
}