---
title: Badger Divide (Reversed)
date: 2026-05-08
categories:
- cycling
- travel
tags:
- bikepacking
- scotland
- adventure
published: true
description: The Badger Divide in reverse (Glasgow to Inverness) — a multi-day bikepacking trip across the Scottish Highlands in April with friends.
---

## The Route

The Badger Divide runs from Glasgow to Inverness (or reverse, as I did it) through some of the most remote and beautiful landscapes in Scotland.

<div id="route-map-badger" class="not-prose rounded-lg overflow-hidden" style="height:450px;width:100%;"></div>
<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"/>
<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
<script>
(function(){fetch('/api/routes/badger-divide/coords').then(function(r){return r.json()}).then(function(coords){var map=L.map('route-map-badger').setView([55.94,-4.31],8);L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',{attribution:'&copy; OpenStreetMap contributors',maxZoom:18}).addTo(map);var pts=coords.map(function(c){return[c.lat,c.lon]});var poly=L.polyline(pts,{color:'#e74c3c',weight:3}).addTo(map);map.fitBounds(poly.getBounds())})})();
</script>

## Weather

## Kit list

## Where we stayed

## Where we got food

