var SearchTitles = document.getElementsByClassName("search-collapsible");
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

document.getElementById("items").innerHTML = "Found " + SearchTitles.length + " results."

function getQueryVariable(variable) {
    var query = window.location.search.substring(1);
    var vars = query.split("&");
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split("=");
        if (pair[0] == variable) { return pair[1]; }
    }
    return (false);
}

var Query = getQueryVariable("q")
document.getElementById("input-query").value = Query; 