const likeButtons = document.querySelectorAll('.like-button');
const dislikeButtons = document.querySelectorAll('.dislike-button');
const commentBtn = document.querySelectorAll('.comment-btn')
const commentSct = document.querySelector('.commentSection')
const filter = document.querySelector('.filter')
const filterBox = document.querySelector('.filterBox')
const filterBtn = document.querySelector('.filterBtn')
const burgerBox = document.querySelector('.burgerBox')
const clearFilter = document.querySelector('.clearFilter')
const fitlerByLikes = document.querySelector('.filterByLikes')
const fitlerByDates = document.querySelector('.filterByDate')
const myLikes = document.querySelector('.myLikes')
const myPosts = document.querySelector('.myPosts')

let bool = false
let burgerBool = false

function clearFilters() {
  window.location.href = "https://localhost/"
}

function deleteCookies() {
  console.log("hi")
  var cookies = document.cookie.split(";");

  // Loop through the array and set each cookie's expiration date to a date in the past
  for (var i = 0; i < cookies.length; i++) {
    var cookie = cookies[i].trim(); // Remove leading and trailing spaces
    var name = cookie.slice(0, cookie.indexOf("=")); // Extract the cookie name
    document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT";
  }
  location.reload();
}

function showBurger() {
  if (burgerBool == false) {
    burgerBool = true
    setTimeout(() => {
      burgerBox.style.visibility = "visible"
    }, 200);
    burgerBox.style.animation = "slide 1s"
  } else {
    burgerBool = false
    setTimeout(() => {
      burgerBox.style.visibility = "hidden"
    }, 600);
    burgerBox.style.animation = "slideOut 1s"
  }
}

function showBarre() {
  if (bool == false) {
    bool = true
    setTimeout(() => {
      myLikes.style.visibility = 'visible'
      myPosts.style.visibility = 'visible'
      fitlerByLikes.style.visibility = "visible"
      fitlerByDates.style.visibility = "visible"
      clearFilter.style.visibility = "visible"
      filterBox.style.visibility = "visible"
      filterBtn.style.visibility = "visible"
    }, 200);
    setTimeout(() => {
      myLikes.style.animation = 'slideLeft 1s'
    }, 200)
    setTimeout(() => {
      myPosts.style.animation = 'slideLeft 1s'
    }, 150)
    setTimeout(() => {
      fitlerByLikes.style.animation = "slideLeft 1s"
    }, 100)
    fitlerByDates.style.animation = "slideLeft 1s"
    clearFilter.style.animation = "slide 1s"
    filterBtn.style.animation = "slide 1s"
    filterBox.style.animation = "slide 1s"
    console.log("yo")
  } else {
    bool = false
    console.log("bite")
    setTimeout(() => {
      myLikes.style.visibility = 'hidden'
      myPosts.style.visibility = 'hidden'
      fitlerByLikes.style.visibility = "hidden"
      fitlerByDates.style.visibility = "hidden"
      clearFilter.style.visibility = "hidden"
      filterBox.style.visibility = "hidden"
      filterBtn.style.visibility = "hidden"
    }, 600);
    myLikes.style.animation = 'slideRight 1s'
    myPosts.style.animation = 'slideRight 1s'
    fitlerByLikes.style.animation = "slideRight 1s"
    fitlerByDates.style.animation = "slideRight 1s"
    clearFilter.style.animation = "slideOut 1s"
    filterBtn.style.animation = "slideOut 1s"
    filterBox.style.animation = "slideOut 1s"

  }
}

likeButtons.forEach(function (likeButton) {
  likeButton.addEventListener('click', function () {
    setTimeout(function () {
      location.reload();
    }, 100);
  });
});

dislikeButtons.forEach(function (dislikeButton) {
  dislikeButton.addEventListener('click', function () {
    setTimeout(function () {
      location.reload();
    }, 100);
  });
});

function reverse() {
  var myDiv = document.querySelector(".theRest"); // Get the div element
  var myArray = Array.prototype.slice.call(myDiv.childNodes);
  console.log(myArray) // Convert the div content to an array
  myArray = myArray.filter(function (element) { // Filter out elements with class "restTitle"
    return element.nodeType === 1 && !element.classList.contains("restTitle");
  });
  myArray.reverse(); // Reverse the order of the elements in the array
  myArray.forEach(function (element) { // Add the reversed elements back into the div in reverse order
    myDiv.appendChild(element);
  });
}

function filterByLikes() {
  var urlWithLikes = "https://localhost/?likes=true";
  window.location.href = urlWithLikes;
}

function filterUserLikes() {
  var urlWithLikes = "https://localhost/?user_likes=true";
  window.location.href = urlWithLikes;
}

function filterUserPosts() {
  var urlWithLikes = "https://localhost/?user_posts=true";
  window.location.href = urlWithLikes;
}

function setFileName(url) {
  const fileName = url.split(/(\\|\/)/g).pop()
  let description = document.querySelector('.description').innerText = fileName
  console.log(url, fileName)
}