
var map
var plotlayers = {}

var selmarker = L.icon({
    iconUrl : '/script/leaflet/images/marker-icon-sel.png'
})

function create_place_div(place, layerid) {
    d = document.createElement("div")
    d.setAttribute("class", "place")
    d.setAttribute("layerid", layerid)
    a = document.createElement("a")
    a.setAttribute("href", "/places/" + place.Name)
    a.appendChild(document.createTextNode(place.Name))
    d.appendChild(a)
    return d
}

function create_availability_div(item, layerid) {
    d = document.createElement("div")
    d.setAttribute("class", "place")
    d.setAttribute("layerid", layerid)
    d.appendChild(document.createTextNode(item.Description))
    return d    
}

function fetch_locations() {
    var b = map.getBounds()
    
    $.getJSON('/api/stuff_at',
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
                      var item = json[i]
                      var place
                      var layerid = item.Type + item.Id
                      if (item['Type'] == 'place') {
                          place = item
                          pl.append(create_place_div(item, layerid))
                      } else {
                          /* BUG: should only show places on the map. This
                             way places will be added more than one time
                          */
                          place = item['Place']
                          pl.append(create_availability_div(item, layerid))
                      }

                      // 2: add to map
                      if (layerid in oldlayers) {
                          plotlayers[layerid] = oldlayers[layerid]
                          delete oldlayers[layerid]
                      } else {
                          // only new listings are added here, so
                          // we could display an effect
                          console.log(place)
                          m = new L.Marker(new L.LatLng(place.Lat, place.Long))
                          map.addLayer(m)
                          m.bindPopup(place.Name)
                          plotlayers[layerid] = m
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
        layerid = $(this).context.getAttribute("layerid")
        m = plotlayers[layerid]
        m.setIcon(selmarker)
    })
    $("#places").on("mouseout", ".place", function(e) {
        layerid = $(this).context.getAttribute("layerid")
        m = plotlayers[layerid]
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

        
