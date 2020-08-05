function validate_form() {
    var i = 0;
    i = 0;

    var button = document.getElementById('submit-button');

    var password = document.getElementById('password').value
    var RegExp_password = /^(?=.*\d)(?=.*[~!@#$%^&*()_\-+=|\\{}[\]:;<>?/])(?=.*[A-Z])(?=.*[a-z])\S{8,60}$/;
    var validPassword = RegExp_password.test(password);

    if (validPassword) {
        document.getElementById('message_pw_complex').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightgreen\'>Password is strong</span>';
        i += 1
    } else {
        document.getElementById('message_pw_complex').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightcoral\'>Password is not complex enough</span>';
    }

    if (document.getElementById('password').value ==
        document.getElementById('password-repeat').value) {
        document.getElementById('message_pw_repeat').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightgreen\'>Password are matching</span>';
        i += 1
    } else {
        document.getElementById('message_pw_repeat').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightcoral\'>Passwords are not matching</span>';
    }

    var name = document.getElementById('name').value
    var RegExp_name = /^([A-Z][A-Za-z.'\-]+) (?:([A-Z][A-Za-z.'\-]+) )?([A-Z][A-Za-z.'\-]+)$/;
    var validName = RegExp_name.test(name);

    if (validName) {
        document.getElementById('message_name').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightgreen\'>Valid name</span>';
        i += 1
    } else {
        document.getElementById('message_name').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightcoral\'>Invalid name</span>';
    }

    var email = document.getElementById('email').value
    var RegExp_email = /^[\w.%+\-]+@[\w.\-]+\.[A-Za-z]{2,6}$/;
    var validEmail = RegExp_email.test(email);

    if (validEmail) {
        document.getElementById('message_email').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightgreen\'>Valid E-mail</span>';
        i += 1
    } else {
        document.getElementById('message_email').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightcoral\'>Invalid E-mail</span>';
    }

    var username = document.getElementById('username').value
    var RegExp = /^([A-Za-z0-9_]){1,15}$/;
    var validUsername = RegExp.test(username);

    if (validUsername) {
        document.getElementById('message_username').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightgreen\'>Valid username</span>';
        i += 1
    } else {
        document.getElementById('message_username').innerHTML = '<span style=\'padding: 6px; border-radius: 6px; background-color: lightcoral\'>Invalid username</span>';
    }
    console.log(i)
    if (i == 5) {
        button.disabled = false;
    } else {
        button.disabled = true;
    }
}