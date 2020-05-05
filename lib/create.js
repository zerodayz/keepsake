function validate_form() {
    var i = 0;
    i = 0;

    var button = document.getElementById('submit-button');

    var password = document.getElementById('password').value
    var RegExp_password = /^(?=.*\d)(?=.*[~!@#$%^&*()_\-+=|\\{}[\]:;<>?/])(?=.*[A-Z])(?=.*[a-z])\S{8,60}$/;
    var validPassword = RegExp_password.test(password);

    if (validPassword) {
        document.getElementById('message_pw_complex').innerHTML = 'Password is strong';
        i += 1
    } else {
        document.getElementById('message_pw_complex').innerHTML = 'Password is not complex enough';
    }

    if (document.getElementById('password').value ==
        document.getElementById('password-repeat').value) {
        document.getElementById('message_pw_repeat').innerHTML = 'Password are matching';
        i += 1
    } else {
        document.getElementById('message_pw_repeat').innerHTML = 'Passwords are not matching';
    }

    var name = document.getElementById('name').value
    var RegExp_name = /^([A-Z][A-Za-z.'\-]+) (?:([A-Z][A-Za-z.'\-]+) )?([A-Z][A-Za-z.'\-]+)$/;
    var validName = RegExp_name.test(name);

    if (validName) {
        document.getElementById('message_name').innerHTML = 'Valid name';
        i += 1
    } else {
        document.getElementById('message_name').innerHTML = 'Invalid name';
    }

    var email = document.getElementById('email').value
    var RegExp_email = /^[\w.%+\-]+@[\w.\-]+\.[A-Za-z]{2,6}$/;
    var validEmail = RegExp_email.test(email);

    if (validEmail) {
        document.getElementById('message_email').innerHTML = 'Valid E-mail';
        i += 1
    } else {
        document.getElementById('message_email').innerHTML = 'Invalid E-mail';
    }

    var username = document.getElementById('username').value
    var RegExp = /^([A-Za-z0-9_]){1,15}$/;
    var validUsername = RegExp.test(username);

    if (validUsername) {
        document.getElementById('message_username').innerHTML = 'Valid username';
        i += 1
    } else {
        document.getElementById('message_username').innerHTML = 'Invalid username';
    }
    console.log(i)
    if (i == 5) {
        button.disabled = false;
    } else {
        button.disabled = true;
    }
}