var SearchTitles = document.getElementsByClassName("search-collapsible");
var i;
console.log(SearchTitles.length)
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