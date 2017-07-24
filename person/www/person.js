"use strict";

document.addEventListener("DOMContentLoaded", function(e) {
	ready();
});

function ready() {
	document.getElementById("person").addEventListener("submit", function(e) {
		e.preventDefault();
		console.log(e.target.getElementsByTagName("input"));
	});
}
