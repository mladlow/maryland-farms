// © OpenStreetMap contributors
// All map data and geocoding is available under the Open Database License
// (https://www.openstreetmap.org/copyright).

function handleMarkerClick(farm) {
    let html = `
        <h1>${farm["Name"]}</h1>
        <p>${farm["Address"]}</p>
    `;
    let phone = farm["Phone"];
    if (phone) {
        html += `<p>${phone}</p>`;
    }
    let website = farm["Website"];
    if (website) {
        if (!website.startsWith("http")) {
            website = `https://${website}`;
        }
        html += `<a target="_blank" href="${website}">${website}</a>`;
    }

    infoPane.innerHTML = html;
}

var infoPane = document.getElementById("info");
var farmMap = L.map('map').setView([38.899, -77.084], 8);
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attributions: '&copy; <a href="https://openstreetmap.org/copyright">OpenStreetMap contributors</a>'
}).addTo(farmMap);

var markers = L.markerClusterGroup();

for (let i = 0; i < farms.length; i++) {
    let farm = farms[i];
    let title = farm["Name"];
    var marker = L.marker(new L.LatLng(farm["Lat"], farm["Lng"]), {title: title})
        .on('click', function(e) {
            handleMarkerClick(farm);
        });
    marker.bindPopup(title);
    markers.addLayer(marker);
}

farmMap.addLayer(markers);
