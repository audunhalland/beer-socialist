
var map;

function initmap() {
    map = new L.map('beermap')

    var osmUrl = 'http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png';
    var osmAttrib = 'Map data © <a href="http://openstreetmap.org">OpenStreetMap</a> contributors'

    var osm = new L.TileLayer(osmUrl, {minZoom : 8, maxZoom : 20, attribution : osmAttrib});
    //map.setView(new L.LatLng(51.3, 0.7), 9);
    map.addLayer(osm)
}

function formatdate(unix) {
    date = new Date(unix * 1000)
    return date.toDateString()
}

window.onload = function() {
    initmap()

    /*
    window.setInterval(function() {
        console.log("hehe")
        $.getJSON("/list",
                  "",
                  function(json) {
                      console.log(json)
                      for (var i = 0; i < json.length; i++) {
                          console.log(json[i])
                          $("#dynamiclist").append("<li>" + json[i].Name + "</li>")
                      }
                  })
    }, 1000)
    */

    $.getJSON('/json/homecoord', '',
              function(json) {
                  map.setView(new L.LatLng(json[0], json[1]), 11)
              })

    $.getJSON('/api/availability', '',
              function(json) {
                  for (var i = 0; i < json.length; i++) {
                      p = json[i].Period;
                      $("#availability").append("<li>" + json[i].Description + " " +
                                                formatdate(p.Start) + " - " +
                                                formatdate(p.End) +
                                                "</li>")
                  }
              })
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

        
