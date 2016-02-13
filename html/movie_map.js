// Markers and other controls are stored in the variables

// variable for map
var map;

// marker, info window variables
var anchor_marker;
var pop_marker;

var infowindow;
var m_infowindow;

var markers = [];


var geocoder;

// variable for directions
var directionsService;
var directionsDisplay;
var defaultLocation;


// cache variables for the current values
var curLocation;
var curMovie;
// All the icons are from http://maps.google.com/mapfiles/ms/icons/ 
var anchor_icon = 'blue-dot.png'
var pop_icon = 'red-dot.png'
var movie_icon = 'movie_camera.jpg'


// Initialize the main map
function initializeMap() {
	
	// Create the SF map with the center of city as initial value
	defaultLocation = new google.maps.LatLng(37.78, -122.454150)
	curLocation = new google.maps.LatLng(37.78, -122.454150)
	map = new google.maps.Map(document.getElementById('gmap-canvas'), {
		zoom: 13,
		center: defaultLocation,
		mapTypeId: google.maps.MapTypeId.ROADMAP
	});

	// create all the google map services
	geocoder = new google.maps.Geocoder;
	directionsService = new google.maps.DirectionsService;
	directionsDisplay = new google.maps.DirectionsRenderer;

	directionsDisplay.setMap(map);
	directionsDisplay.setOptions( { suppressMarkers: true, preserveViewport: true } );

	// default location infowindow
	infowindow = new google.maps.InfoWindow;
	infowindow.setContent('San Francisco');
	
	// infowindow for movie locations ... 
	// reusing this m_infowindow instead of creating for every click on pin
	m_infowindow = new google.maps.InfoWindow;

	curMovie = ""
	// Create the default markers for anchor_marker and user_selected marker
	anchor_marker = new google.maps.Marker({
		map: map,
		position: defaultLocation,
		icon: anchor_icon,
		draggable: true
	});

	pop_marker = new google.maps.Marker({
		map:map,
		animation: google.maps.Animation.DROP,
		icon: pop_icon,
		clickable: false,
		zIndex: 10,
		optimized: false
	});

	google.maps.event.addListener(anchor_marker, 'dragend', function(event) {
		placeMarkerAndPanTo(event.latLng, map);
	});

	google.maps.event.addListener(anchor_marker, 'click', function() {
		infowindow.open(anchor_marker.get('map'), anchor_marker);
	});

	google.maps.event.addListener (map, 'click', function(event) {
  		clearSearchData();
	});

}

function setDefaultLocation() {
	map.setCenter(defaultLocation)
	setCurLocation(defaultLocation.lat, defaultLocation.lng)
	placeMarkerAndPanTo(defaultLocation, map)
}

function setCurLocation(lat, lng) {
	//curLocation.lat = lat;
	//curLocation.lng = lng;
	curLocation = new google.maps.LatLng(lat,lng);
}

function updateCurrentLocation() {
	
	if (navigator.geolocation) {
		navigator.geolocation.getCurrentPosition(function(position) {
			var pos  = new google.maps.LatLng(position.coords.latitude,position.coords.longitude);
			map.setCenter(pos);
			placeMarkerAndPanTo(pos,map)
			setCurLocation(position.coords.latitude,position.coords.longitude);
		}, function() {
			handleLocationError(true, infoWindow, map.getCenter());
		});
	} else {
    	// Browser doesn't support Geolocation
    	handleLocationError(false, infoWindow, map.getCenter());
    }

}

function handleLocationError(browserHasGeolocation, infoWindow, pos) {
  infoWindow.setPosition(pos);
  infoWindow.setContent(browserHasGeolocation ?
                        'Error: The Geolocation service failed.' :
                        'Error: Your browser doesn\'t support geolocation.');
}

function getMovieLocations(moviename) {
	curMovie = moviename
	var searchurl = "/Locations/"+ moviename;

	$.ajax({
		type: 'GET',
		url: searchurl,
		data: {},
		dataType: 'json',
		success: function(data) 
		{ 
			populateMovieLocations(moviename,data) 
		},
		error: function() { alert('Error occured while getting movie locations !!!'); }
	});

}

function clearSearchData() {
	document.getElementById('srch-input').value = "";
	document.getElementById('srch-results').innerHTML = "";
}

function setCurrentMovie(moviename) {
	document.getElementById('srch-input').value = moviename;
	document.getElementById('srch-results').innerHTML = "";
	getMovieLocations(moviename);
}

$(document).ready(function () {
	
    $("#srch-input").on("input", function () {
        var options = {};
        var search_word  = $("#srch-input").val()
        options.url = "/Movies/"+search_word;
        options.type = "GET";
        options.data = {};
        options.dataType = "json";
        
        options.success = function (data) {
            populateMovieNames(data,search_word);
        };
        options.error = function(a,b) {
        	document.getElementById('srch-input').value = moviename;
			document.getElementById('srch-results').innerHTML = "";
        };
        
        $.ajax(options);
      
    });
    

});

function populateMovieNames(data, search_word) {

	var ih = "";
	if(data && data.length) {
		for(var i=0;i<data.length;i++) {
			ih = ih + "<tr onclick=" + '\"setCurrentMovie(\'' + data[i].Moviename + '\')\"' +'>'
			ih = ih + "<td>" + data[i].Moviename + "</td></tr>";
		}
		document.getElementById('srch-results').innerHTML = ih;
	} else {
		document.getElementById('srch-results').innerHTML = ih;
		document.getElementById('srch-input').value =search_word;
	}

}

function populateMovieLocations (moviename, data) {
	deleteMarkers()
	directionsDisplay.setDirections({routes: []});

	document.getElementById('results').innerHTML = "";
	document.getElementById('moviename').innerHTML = data.Title;
	document.getElementById('director').innerHTML = data.Director;
	document.getElementById('released').innerHTML = data.Released;

	var actors = data.Actors[0]
	for(var i=1;i<data.Actors.length;i++) {
		actors = actors + ", " + data.Actors[i]
	}
	document.getElementById('actors').innerHTML = actors;


	var bounds = new google.maps.LatLngBounds()
	bounds.extend(curLocation)

	var innerhtml ="<tr><td></td></tr>";
	var locdata = data.Spots
	for(var i=0;i<locdata.length;i++) {
		var id = "";
		id = id + i;
		var loc = { lat: locdata[i].Latlng.Lat, lng: locdata[i].Latlng.Lng};

		if (loc.lat == 0.0 && loc.lng == 0.0) 
			continue;

		var tmp = new google.maps.LatLng(loc.lat, loc.lng)
		bounds.extend(tmp);
		addMovieLocationMarker(loc, id, moviename, locdata[i].Address,locdata[i].Funfact);

		innerhtml = innerhtml + "<tr><td>"  
		innerhtml = innerhtml + "<a onclick= " + 'showInfoWindow('+i+')' +' onmouseover= popMarker(' + loc.lat + ',' + loc.lng+ ')>' + locdata[i].Address + "</a><br>"
		innerhtml = innerhtml + "<a style=\"color:darkolivegreen\;\" onclick= " + 'findDirections(' + loc.lat + ',' + loc.lng + ',\"DRIVING\"'+ ',' + i +')>' + '<b>Directions</b></a>'
		innerhtml = innerhtml + "</td></tr>" ;
	}
	map.fitBounds(bounds);
	document.getElementById('results').innerHTML = innerhtml;
}
// Function for mouse click on a row of the table element
function showInfoWindow(id) {
	google.maps.event.trigger(markers[id], 'click');
}

function popMarker(lat, lng) {
	var loc = new google.maps.LatLng(lat, lng) 
	pop_marker.setPosition(loc);
	pop_marker.setVisible(true);
	m_infowindow.close();
	directionsDisplay.setDirections({routes: []});
}

// Add marker for movie location

function addMovieLocationMarker(location, id, name, address,funfacts) {

	var marker = new google.maps.Marker({
		position: location,
		map: map,
		icon: movie_icon,
		clickable: true
	});

	google.maps.event.addListener(marker, 'click', function() {
		if(funfacts === "")
			m_infowindow.setContent("<b>" + name + "</b>" + "<br>" + address)
		else
			m_infowindow.setContent("<b>" + name + "</b>" + "<br>" + address + "<p><b>Funfact: </b>" + funfacts + "</p>")
		m_infowindow.open(map, marker)
		directionsDisplay.setDirections({routes: []});
	});

	google.maps.event.addListener(marker, 'dblclick', function() {
		m_infowindow.close()
		directionsDisplay.setDirections({routes: []});
	});
	markers.push(marker);

	return marker;
}

function toggleBounce(marker) {
  if (marker.getAnimation() !== null) {
    marker.setAnimation(null);
  } else {
    marker.setAnimation(google.maps.Animation.BOUNCE);
  }
}

// Sets the map on all markers in the array.
function setMapOnAll(map) {
	for (var i = 0; i < markers.length; i++) {
		markers[i].setMap(map);
	}
}

// Removes the markers from the map, but keeps them in the array.
function clearMarkers() {
	setMapOnAll(null);
}

// Shows any markers currently in the array.
function showMarkers() {
	setMapOnAll(map);
}

// Deletes all markers in the array by removing references to them.
function deleteMarkers() {
	directionsDisplay.setDirections({routes:[]})
	pop_marker.setVisible(false)
	clearMarkers();
	markers = [];
}

// Get the json for a given URL
// function GetJson(yourUrl) {
// 	var Httpreq = new XMLHttpRequest(); // a new request
// 	Httpreq.open("GET", yourUrl, false);
// 	Httpreq.send(null);
// 	return Httpreq.responseText;
// }

// Place the anchor_marker and pan to the location
function placeMarkerAndPanTo(latLng, map) {

	geocoder.geocode({
		'location': latLng
	}, function(results, status) {
		if (status === google.maps.GeocoderStatus.OK) {
			if (results[1]) {
				infowindow.setContent(results[1].formatted_address);
			} else {
				window.alert('No results found');
			}
		} else {
			window.alert('Geocoder failed due to: ' + status);
		}
	});

	anchor_marker.setPosition(latLng);
	setCurLocation(latLng.lat(), latLng.lng())
	directionsDisplay.setDirections({routes: []});
}

function findDirections(dlat, dlng, mode, id) {

	popMarker(dlat, dlng);
	var origin_latlng = new google.maps.LatLng(curLocation.lat(), curLocation.lng())
	var dest_latlng = new google.maps.LatLng(dlat, dlng);

	directionsService.route({
		origin: origin_latlng,
		destination: dest_latlng,
		travelMode: google.maps.TravelMode[mode]
	}, function(response, status) {
		if (status === google.maps.DirectionsStatus.OK) {
			directionsDisplay.setDirections(response);
		} else {
			window.alert('Directions request failed due to ' + status);
		}
	});
}

// Initialize the google map on load
google.maps.event.addDomListener(window, 'load', initializeMap);


