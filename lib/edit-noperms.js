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