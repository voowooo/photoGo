console.log('script is works')

function HelloTextChange(){
    document.getElementById('helloText').innerHTML = "Привет, " + document.getElementById('username').value + "!👋";
}

function settingsShowPass(){
document.getElementById('settings-password').style.display = "flex";
document.getElementById('settings-password-button').style.display = "none";
}

function showInputPass(what){
    if(what == 0){
        document.getElementById('password').type = "password";
    } else {
        document.getElementById('password').type = "text";
    }
}

function deleteAccount(){
    document.getElementById('delete-account-div').style.display = "flex";
}

function close(what){
    console.log('close')
    if (what == "delete-account-div"){
        document.getElementById('delete-account-div').style.display = "none";
    }
}



// const allLangs = ["by", "ru"];
// let currentLang = "ru";
// const langButtons = document.querySelectorAll("[data-btn]")

// const settingsPageTexts = {
//     "settings_showPassword": {
//         by: "Ваш пароль!!!:",
//         ru:"Ваш пароль:",
//     },
// }