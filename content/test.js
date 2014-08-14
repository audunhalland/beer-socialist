
var map
var plotlayers = {}
var place_assoc = {}
var avails = {}

var selmarker = L.icon({
    iconUrl : '/script/leaflet/images/marker-icon-sel.png'
})

function add_place(place) {
    // only new listings are added here, so
    // we could display an effect
    m = new L.Marker(new L.LatLng(place.Lat, place.Long))
    map.addLayer(m)
    m.bindPopup(place.Name)
    plotlayers[place.Id] = m
}

function add_avail(avail) {
    var place = avail.Place

    if (place.Id in place_assoc) {
        place_assoc[place.Id].push(avail)
    } else {
        place_assoc[place.Id] = [avail]
    }

    avails[avail.Id] = avail

    var d = document.createElement("div")
    d.setAttribute("class", "avail")
    d.setAttribute("avail_id", avail.Id)
    d.appendChild(document.createTextNode(avail.Participant.Alias + "@" + avail.Place.Name))
    $("#places").append(d)
}

function map_highlight(placeid, enabled) {
    m = plotlayers[placeid]
    if (enabled) {
        m.setIcon(selmarker)
    } else {
        m.setIcon(new L.Icon.Default())
    }
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
                  place_assoc = {}
                  pl = $("#places")
                  pl.empty()

                  for (var i = 0; i < json.length; i++) {
                      var item = json[i]
                      if (item['Type'] == 'place') {
                          if (item.Id in oldlayers) {
                              plotlayers[item.Id] = oldlayers[item.Id]
                              delete oldlayers[item.Id]
                          } else {
                              add_place(item)
                          }
                      } else {
                          add_avail(item)
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


    $("#places").on("mouseover", ".avail", function(e) {
        avail = avails[$(this).context.getAttribute("avail_id")]
        map_highlight(avail.Place.Id, true)
    })
    $("#places").on("mouseout", ".avail", function(e) {
        avail = avails[$(this).context.getAttribute("avail_id")]
        map_highlight(avail.Place.Id, false)
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

        
