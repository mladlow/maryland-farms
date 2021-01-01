// © OpenStreetMap contributors
// All map data and geocoding is available under the Open Database License
// (https://www.openstreetmap.org/copyright).
var farmMap = L.map('map').setView([38.899, -77.084], 8);
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attributions: '&copy; <a href="https://openstreetmap.org/copyright">OpenStreetMap contributors</a>'
}).addTo(farmMap);

var markers = L.markerClusterGroup();

for (var i = 0; i < farms.length; i++) {
    var farm = farms[i];
    var title = farm["Name"];
    var marker = L.marker(new L.LatLng(farm["Lat"], farm["Lng"]), {title: title});
    marker.bindPopup(title);
    markers.addLayer(marker);
}

farmMap.addLayer(markers);
