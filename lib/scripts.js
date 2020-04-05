function toggle(elementId) {
    var ele = document.getElementById(elementId);
    if (ele.style.display == "block") {
        ele.style.display = "none";
    }
    else {
        ele.style.display = "block";
    }
} 