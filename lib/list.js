
var items = document.getElementById("items");
var allCheckboxes = document.querySelectorAll('input[type=checkbox]');
var allCategories = Array.from(document.querySelectorAll('.category'));
var checked = {};

function getChecked(name) {
    checked[name] = Array.from(document.querySelectorAll('input[name=' + name + ']:checked')).map(function (el) {
        return el.value;
    });
}

function setVisibility() {
    allCategories.map(function (el) {
        var tags = checked.tags.length ? _.intersection(Array.from(el.classList), checked.tags).length : true;
        if (tags) {
            el.parentElement.style.display = 'table-row';
        } else {
            el.parentElement.style.display = 'none';
        }
    });
}

function toggleCheckbox(e) {
    getChecked(e.target.name);
    setVisibility();
}

Array.prototype.forEach.call(allCheckboxes, function (el) {
    el.addEventListener('change', toggleCheckbox);
});

getChecked('tags');


const classA = Array.from(document.getElementsByClassName("search-collapsible"))
    ,classB = Array.from(document.getElementsByClassName("search-no-collapsible"))
    ,classC = Array.from(document.getElementsByClassName("category"))
    ,SearchTitles = Array.from(new Set(classA.concat(classB)))
    ,Items = Array.from(classC)
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

if (items) {
    items.innerHTML = "Found " + Items.length + " pages."
}