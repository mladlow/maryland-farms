// © OpenStreetMap contributors
// All map data and geocoding is available under the Open Database License
// (https://www.openstreetmap.org/copyright).
var farmMap = L.map('map').setView([38.899, -77.084], 8);
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attributions: '&copy; <a href="https://openstreetmap.org/copyright">OpenStreetMap contributors</a>'
}).addTo(farmMap);

var marker = L.marker(["39", "-77.105"]).addTo(farmMap);
