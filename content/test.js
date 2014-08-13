
var map
var plotlayers = {}

var selmarker = L.icon({
    iconUrl : '/script/leaflet/images/marker-icon-sel.png'
})

function create_place_div(p) {
    d = document.createElement("div")
    d.setAttribute("class", "place")
    d.setAttribute("placeid", p.Id)
    a = document.createElement("a")
    a.setAttribute("href", "/places/" + p.Name)
    a.appendChild(document.createTextNode(p.Name))
    d.appendChild(a)
    return d
}

function fetch_locations() {
    var b = map.getBounds()
    
    $.getJSON('/api/places',
              {"minlat"  : b.getSouthWest().lat,
               "minlong" : b.getSouthWest().lng,
               "maxlat"  : b.getNorthEast().lat,
               "maxlong" : b.getNorthEast().lng
              },
              function(json) {
                  oldlayers = plotlayers
                  plotlayers = {}
                  pl = $("#places")
                  pl.empty()

                  for (var i = 0; i < json.length; i++) {
                      p = json[i]

                      // 1: add to place list
                      pl.append(create_place_div(p))

                      // 2: add to map
                      if (p.Id in oldlayers) {
                          plotlayers[p.Id] = oldlayers[p.Id]
                          delete oldlayers[p.Id]
                      } else {
                          // only new listings are added here, so
                          // we could display an effect
                          m = new L.Marker(new L.LatLng(p.Lat, p.Long))
                          map.addLayer(m)
                          m.bindPopup(p.Name)
                          plotlayers[p.Id] = m
                      }
                  }

                  for (var key in oldlayers) {
                      map.removeLayer(oldlayers[key])
                  }
              }
              )
}

function initmap() {
    map = new L.map('beermap')

    var osmUrl = 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png';
    var osmAttrib = 'Map data © <a href="https://openstreetmap.org">OpenStreetMap</a> contributors'

    var osm = new L.TileLayer(osmUrl, {minZoom : 8, maxZoom : 20, attribution : osmAttrib});
    map.addLayer(osm)
    map.on('moveend', function(e) { fetch_locations() })
}

function formatdate(unix) {
    date = new Date(unix * 1000)
    return date.toDateString()
}

window.onload = function() {
    initmap()

    $.getJSON('/api/userpref?q=homelat&q=homelong', '',
              function(json) {
                  map.setView(new L.LatLng(json["homelat"], json["homelong"]), 11)
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


    $("#places").on("mouseover", ".place", function(e) {
        placeid = $(this).context.getAttribute("placeid")
        m = plotlayers[placeid]
        m.setIcon(selmarker)
    })
    $("#places").on("mouseout", ".place", function(e) {
        placeid = $(this).context.getAttribute("placeid")
        m = plotlayers[placeid]
        m.setIcon(new L.Icon.Default())
    })

    fetch_locations()

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

        
