console.log('script is works')
// import axios from 'axios';

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
    if(document.getElementById('password')){
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

function preWatchNewPhoto(){
    // document.getElementById('NewPhoto').src = document.getElementById('addPhoto').value;
    
    let fileInput = document.getElementById('addPhoto');
    let file = fileInput.files[0];

    if (file) {
        let reader = new FileReader();
        reader.onload = function(e) {
            document.getElementById('NewPhoto').src = e.target.result;
        }
        reader.readAsDataURL(file);
    }
}



function checkHeaderY(){
    const header = document.getElementById('header');
    const SH = document.getElementById('SH');
    let pageOffset = window.pageYOffset;

    if(pageOffset > 30){
        // SH.style.pointerEvents = "all";
        // SH.style.opacity = "100%";
        header.classList.add('header-wrapp');
        let scrollDir = '';
        if(window.location.href.includes("/profile") && pageOffset < 200){
            const profileElement = document.getElementById('profile');

            // Получаем позицию элемента profile относительно документа и его высоту
            const profileBottom = profileElement.getBoundingClientRect().bottom + window.pageYOffset;

            // Прокручиваем страницу до позиции, где элемент profile будет скрыт (вне видимой области)
            window.scrollTo({
                top: profileBottom - 90,
                behavior: 'smooth' // Плавная прокрутка
            });

            // window.scroll({ 
            //     top: 430,
            //     left: 0,
            //     behavior: 'smooth' // Это как катание на круизном лайнере 🛥️
            // });
        }
    }else{
        // SH.style.opacity = "0%";
        // SH.style.pointerEvents = "none";
        header.classList.remove('header-wrapp');
    }
}

function goup(){
    window.scroll({
        top: 0,
        left: 0,
        behavior: 'smooth' // Это как катание на круизном лайнере 🛥️
    });
}

function show(what){
    if(what == 'settingsColor'){
        document.getElementById('settingsColorWindow').style.display = 'flex';
    } else if(what == 'shareUserCard'){
        document.getElementById('shareUserCard').style.display = 'flex';
    }
}

function closeShareUserCard(){
    document.getElementById('shareUserCard').style.display = 'none';
}

function close(what){
    console.log('close');
    if(what == 'photoDIV'){
        document.getElementById('photoDIV').style.display = 'none';
    }
}

function closePhotoDiv() {
    console.log('close');
    document.getElementById('photoDIV').style.display = 'none';
    // window.location.href.split('#')[0];
    // window.location = window.location.href.split('#')[0];
    // window.location.href = window.location.href.split('#')[0];
    // console.log(window.location.href.split('#')[0])
    history.replaceState(null, null, window.location.href.split('#')[0]);
}

function changecolor(color){
    document.getElementById('settingsColorPickerUserCard').style.backgroundColor = 'rgb(' + color + ')';
    document.getElementById('pickedColor').value = color;
}

const hexToRgb = hex => {
    let [r, g, b] = hex.match(/\w\w/g).map(x => parseInt(x, 16));
    return `${r}, ${g}, ${b}`;
};

function changeCustomColor(){
    document.getElementById('colorPickerLabel').style.backgroundColor = document.getElementById('colorPickerInput').value;
    document.getElementById('settingsColorPickerUserCard').style.backgroundColor = document.getElementById('colorPickerInput').value;
    document.getElementById('pickedColor').value = hexToRgb(document.getElementById('colorPickerInput').value);
}

window.addEventListener('popstate', function() {
    showFullPhoto();
});

function shareUserCardDirection(){
    const checkbox = document.getElementById('shareUserCardDirectionCheckbox');
    const customCheckbox = document.getElementById('shareUserCardDirToggleDivIn');
    if (checkbox.checked){
        customCheckbox.classList.add("shareUserCard-dir-toggle-div-in-checked");
    } else {
        customCheckbox.classList.remove("shareUserCard-dir-toggle-div-in-checked");
    }
}


// function showFullPhoto(){
//     console.log("SFP")
//     if(window.location.href.includes("/photo")){
//         console.log("SFP if")
//         let imgSrc = window.location.href.substring(window.location.href.indexOf('#') + 1);
//         // const url = "http://localhost:8080/profile/46#/photo/27";
//         const url = window.location.href;
//         const regex = /photo\/(\d+)/;
//         const matches = url.match(regex);

//         let PhotoId = matches[1]

//         document.getElementById('PhotoIdInput').value = PhotoId;

//         if (matches && matches[1]) {
//             console.log("ID фото:", PhotoId);
//         } else {
//             console.log("ID фото не найден");
//         }
//         console.log(imgSrc)
//         document.getElementById('photoDIV').style.display = "flex";
//         document.getElementById('photoIMG').src = imgSrc;
//     }
    
// }




function FullPhoto(imgData) {
    console.log("Полное изображение:", imgData);
    let newUrl = 'http://localhost:8080/fullphoto' + imgData;

    let comments = "";

    const parts = newUrl.split('/');
    const PhotoId = parts[parts.length - 1];

    axios.get(newUrl)
        .then(response => {
            comments = response.data; // обработка полученных данных

            console.log("comments");
            console.log(comments);
            console.log("comments");

            const AllCommsDiv = document.getElementById("AllCommsDiv");

            AllCommsDiv.innerHTML = '';

            for(let i = 0; i < comments.length; i++){
                var commentDiv = document.createElement("div");
                commentDiv.className = 'comm-div';
            
                let commentOwner = document.createElement("h1");
                commentOwner.innerHTML = comments[i].Owner;
            
                let commentText = document.createElement("h2");
                commentText.innerHTML = comments[i].Text;
            
                // Добавьте commentOwner и commentText в commentDiv
                commentDiv.appendChild(commentOwner); // Используйте appendChild вместо присваивания
                commentDiv.appendChild(commentText); // Добавьте текст комментария
            
                // Добавьте commentDiv в photoBodyBlock
                AllCommsDiv.appendChild(commentDiv);
            }


            // console.log(response.data);
        })
        .catch(error => {
            console.error('There was an error!', error);
        });

    document.getElementById('photoDIV').style.display = "flex";
    document.getElementById('photoIMG').src = imgData;
    document.getElementById('PhotoIdInput').value = PhotoId;
    // AllCommsDiv.scrollTo({
    //     top: 0,
    //     behavior: 'smooth' // Плавная прокрутка
    // });;
}


function showFullSubs(){
    document.getElementById('AllSubs').style.display = "flex";
    document.getElementById('AllSubsBodyColumn').scroll = 0;
}

function showFullFollowers(){
    document.getElementById('AllFollowers').style.display = "flex";
    document.getElementById('AllSubsBodyColumn').scroll = 0;
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

// function show(what){
//     if(what == 'SHUSS'){
//         document.getElementById('SHUSS').style.display = "";
//     }
// }

// function hide(what){
//     if(what == 'SHUSS'){
//         setTimeout(
//             () => {
//                 document.getElementById('SHUSS').style.display = "none";
//             },
//             0.5 * 1000
//         );
//     }
// }