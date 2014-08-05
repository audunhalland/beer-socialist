
var map;

function initmap() {
    map = new L.map('beermap')

    var osmUrl = 'http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png';
    var osmAttrib = 'Map data © <a href="http://openstreetmap.org">OpenStreetMap</a> contributors'

    var osm = new L.TileLayer(osmUrl, {minZoom : 8, maxZoom : 20, attribution : osmAttrib});
    map.addLayer(osm)
}

function formatdate(unix) {
    date = new Date(unix * 1000)
    return date.toDateString()
}

window.onload = function() {
    initmap()

    $.getJSON('/json/homecoord', '',
              function(json) {
                  map.setView(new L.LatLng(json[0], json[1]), 11)
              })

    $.getJSON('/api/availability', '',
              function(json) {
                  av = $("#availability")
                  for (var i = 0; i < json.length; i++) {
                      p = json[i].Period;
                      li = document.createElement("li")
                      li.appendChild(document.createTextNode(
                          json[i].Description + " " +
                              formatdate(p.Start) + " - " +
                              formatdate(p.End)))
                      av.append(li)
                  }
              })

    // populate the map
    $.getJSON('/api/places', '',
              function(json) {
                  for (var i = 0; i < json.length; i++) {
                      p = json[i]
                      m = new L.Marker(new L.LatLng(p.Lat, p.Long))
                      map.addLayer(m)
                      m.bindPopup(p.Name)
                  }
              }
              )

    $("#placesearch").autocomplete({serviceUrl: "/api/placesearch"})
}

function fbStatusChanged(response) {
    console.log('fb status changed');
    console.log(response);
}

function checkFbLoginState() {
    FB.getLoginStatus(fbStatusChanged);
}

window.fbAsyncInit = function() {
    FB.init({
        appId   : '{{.FacebookAppid}}',
        cookie  : true,
        xfbml   : true,
        version : 'v2.0'
    });

    FB.getLoginStatus(checkFbLoginState);
}

        
