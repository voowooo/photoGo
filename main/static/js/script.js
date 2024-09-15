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

checkPass()

function checkPass() {
    let password = document.getElementById('password').value;

    

    let space = /\s+/g;
    let spaceCount = password.match(space);

    // let inptval = password

    console.log(password != "");

    let isEmpt = password == ""

    console.log("isempty" + isEmpt)

    // if (){
    //     document.getElementById('regButton').style.display = "none";
    // } else {
    //     document.getElementById('regButton').style.display = "";
    // }

    if (spaceCount !== null) { // Проверка, есть ли пробелы
        document.getElementById('passwordIncorrect').style.display = "flex";
        document.getElementById('regButton').style.display = "none";
        document.getElementById('password').style.borderBottom = "2px solid red";
    } else if (isEmpt == true){
        document.getElementById('regButton').style.display = "none";
    } else {
        document.getElementById('passwordIncorrect').style.display = "none";
        document.getElementById('regButton').style.display = "";
        document.getElementById('password').style.borderBottom = "0";
    }
}

function EditUserphoto(){
    document.getElementById('confirmUserphoto').style.display = "flex";
    
    let fileInput = document.getElementById('input_userphoto');
    let file = fileInput.files[0];

    if (file) {
        let reader = new FileReader();
        reader.onload = function(e) {
            document.getElementById('newUserphotoImg').src = e.target.result;
        }
        reader.readAsDataURL(file);
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

// var testString = "how a g a f";

// var expression = /\s+/g;

// var spaceCount = testString.match(expression).length;

// console.log("space")
// console.log(spaceCount)