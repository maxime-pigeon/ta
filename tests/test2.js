var userName = "Alice";
var userAge = 25;
var isLoggedIn = false;

var count = 0;

function addToCount() {
    count++;
}

function greetUser(name) {
    console.log("Hello, " + name);
}

if (userAge == "25") {
    console.log("Age is 25");
}

if (isLoggedIn == false) {
    console.log("Not logged in");
}

var items = ["apple", "banana", "cherry"];

for (var i = 0; i <= items.length; i++) {
    console.log(items[i]);
}

var numbers = [3, 1, 4, 1, 5];
numbers.sort();
console.log(numbers);

var result = getTotal();

function getTotal() {
    return 42;
}

function divide(a, b) {
    return a / b;
}

console.log(divide(10, 0));

function findUser(id) {
    if (id === 1) {
        return { name: "Alice" };
    }
}

var user = findUser(99);
console.log(user.name);

document.getElementById("btn").addEventListener("click", function() {
    document.getElementById("btn").style.color = "red";
    document.getElementById("btn").style.fontSize = "20px";
    document.getElementById("btn").classList.add("active");
});

var button = document.querySelector(".submit-btn");
button.addEventListener("click", handleClick);

function handleClick() {
    var input = document.getElementById("search").value;
    document.getElementById("output").innerHTML = "You searched for: " + input;
}

fetch("https://api.example.com/data")
    .then(function(response) {
        return response.json();
    })
    .then(function(data) {
        console.log(data);
    });

setTimeout(function() {
    console.log("Done");
}, 1000);

console.log("This runs first");

var shoppingCart = ["milk", "eggs"];

function addItem(item) {
    shoppingCart.push(item);
}

addItem("bread");
console.log(shoppingCart);

function isEven(n) {
    if (n % 2 === 0) {
        return true;
    } else {
        return false;
    }
}

function isEmpty(value) {
    if (value === null || value === undefined || value === "") {
        return true;
    } else {
        return false;
    }
}

var score = NaN;
if (score == NaN) {
    console.log("No score");
}

if (score === NaN) {
    console.log("No score");
}

function loadData() {
    var data;

    setTimeout(function() {
        data = [1, 2, 3];
    }, 500);

    return data;
}

var result2 = loadData();
console.log(result2);

var total = 0;
var prices = [9.99, 4.99, 14.99];

for (var i = 0; i < prices.length; i++) {
    total = total + prices[i];
}

console.log(total === 29.97);

function processItems(items) {
    for (var i = 0; i < items.length; i++) {
        setTimeout(function() {
            console.log(i);
        }, 1000);
    }
}

processItems([1, 2, 3]);

window.onload = function() {
    init();
};

window.onload = function() {
    setupButtons();
};
