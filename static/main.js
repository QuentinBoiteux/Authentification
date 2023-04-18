let form = document.querySelector('.form')
let form2 = document.querySelector('.form-inscription')
let buttons = document.querySelector('.buttons')
let choose = document.querySelector('.choose')
let legende = document.querySelector('.legende')
let legende2 = document.querySelector('.legende2')
let legende3 = document.querySelector('.legende3')
let alreadyAccount = document.querySelector('.alreadyAccount')
let icon = document.querySelector('.Icon')
let windowU = document.querySelector('.windowU')
let burger = document.querySelector('.Burger')
let title = document.querySelector(".text")
let para = document.querySelector('#parametres')
let profil = document.querySelector('#profil')
let deco = document.querySelector("#deco")
let navBar = document.querySelector('.navbar')
let IT = document.querySelector('#IT')
let SPORT = document.querySelector('#SPORT')
let AUTO = document.querySelector('#AUTO')
let MUSIC = document.querySelector('#MUSIC')
let ART = document.querySelector('#ART')
let croix = document.querySelector('.annuler')
let connectTiers = document.querySelector('.connexionTiers')

addEventListener('load', () => {
    connectTiers.style.animation = "slowAppear 1.2s ease-out"
    setTimeout(() => {
        connectTiers.style.animation = "none"
    }, 1200)
})

function onclik() {
    connectTiers.style.animation = 'slowAppear 1.2s ease-out'
    title.classList.add('nani')
    legende.style.display = 'none'
    legende2.style.display = 'block'
    legende2.style.animation = 'Appear 0.5s ease-out'
    form.style.animation = 'Appear 0.5s ease-out'
    form.style.scale = 1
    form.style.display = 'flex'
    buttons.style.display = 'flex'
    choose.style.display = 'none'
}

function inscription() {
    connectTiers.style.animation = 'slowAppear 1.2s ease-out'
    title.classList.add('nani')
    legende.style.display = 'none'
    legende3.style.display = 'block'
    legende3.style.animation = 'Appear 0.5s ease-out'
    form2.style.animation = 'Appear 0.5s ease-out'
    form2.style.scale = 1
    form2.style.display = 'flex'
    buttons.style.display = 'flex'
    choose.style.display = 'none'
}

const back = () => {
    connectTiers.style.animation = 'slowAppear 1.2s ease-out'
    title.classList.remove('nani')
    legende2.style.display = 'none'
    legende3.style.display = 'none'
    legende.style.animation = 'Appear 0.5s ease-out'
    legende.style.display = 'block'
    form.style.display = 'none'
    form2.style.display = 'none'
    buttons.style.display = 'none'
    choose.style.display = 'flex'
    choose.style.animation = 'Appear 0.5s ease-out'
}

alreadyAccount.addEventListener('click', function () {
    connectTiers.style.animation = 'slowAppear 1.2s ease-out'
    legende3.style.display = 'none'
    legende2.style.display = 'block'
    legende2.style.animation = 'Appear 0.5s ease-out'
    form.style.animation = 'Appear 0.5s ease-out'
    form.style.scale = 1
    form.style.display = 'flex'
    form2.style.display = 'none'
    buttons.style.display = 'flex'
    choose.style.display = 'none'
})


