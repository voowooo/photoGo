console.log('script is works')

function HelloTextChange(){
    document.getElementById('helloText').innerHTML = "Привет, " + document.getElementById('username').value + "!👋";
}

function settingsShowPass(){
document.getElementById('settings-password').style.display = "flex";
document.getElementById('settings-password-button').style.display = "none";
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