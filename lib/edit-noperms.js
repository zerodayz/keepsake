var modalHistory = document.getElementById("ModalHistory");

const classA = Array.from(document.getElementsByClassName("search-collapsible"))
    ,classB = Array.from(document.getElementsByClassName("search-no-collapsible"))
    ,SearchTitles = Array.from(new Set(classA.concat(classB)))
var i;

for (i = 0; i < SearchTitles.length; i++) {
    console.log(SearchTitles[i])
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

if (modalHistory) {
    var btnHistory = document.getElementById("ModalButtonHistory");
    var spanHistory = document.getElementsByClassName("close-history")[0];
    btnHistory.onclick = function () {
        modalHistory.style.display = "block";
    }
    spanHistory.onclick = function () {
        modalHistory.style.display = "none";
    }
}

window.onclick = function (event) {
    if (event.target == modalHistory) {
        modalHistory.style.display = "none";
    }
}

function validate_form() {
    var i = 0;
    i = 0;

    var button = document.getElementById('submit-button');
    var comment_title = document.getElementById('comment_title').value
    var RegExp_comment_title = /^[a-zA-Z0-9_]$/;
    var validTitle = RegExp_comment_title.test(comment_title);

    if (validTitle) {
        document.getElementById('message_name').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightgreen\'>Valid name</span>';
        i += 1
    } else {
        document.getElementById('message_name').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightcoral\'>Invalid name</span>';
    }

    var comment_message = document.getElementById('comment_message').value
    var RegExp_comment_message = /^[a-zA-Z0-9_]$/;
    var validMessage = RegExp_comment_message.test(comment_message);

    if (validMessage) {
        document.getElementById('message_name').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightgreen\'>Valid name</span>';
        i += 1
    } else {
        document.getElementById('message_name').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; color: #000000; background-color: lightcoral\'>Invalid name</span>';
    }
    if (i == 2) {
        button.disabled = false;
    } else {
        button.disabled = true;
    }
}