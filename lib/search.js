
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
            el.style.display = 'block';
        } else {
            el.style.display = 'none';
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

if (items) {
    items.innerHTML = "Found " + SearchTitles.length + " results."
}

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
document.getElementById("inputQuery").value = Query;

function postQuery() {
    setTimeout(function() { document.getElementById('submit').click() }, 1000)
};


var txtarea = document.getElementById("inputQuery");
setTimeout(function() { txtarea.focus() }, 500)