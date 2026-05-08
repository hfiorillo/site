---
title: Bikepacking Across Scotland
date: 2026-05-08
categories:
- cycling
- travel
tags:
- bikepacking
- scotland
- adventure
published: true
description: The Badger Divide in reverse — a multi-day bikepacking trip across the Scottish Highlands.
---

## The Route

The Badger Divide runs from Glasgow to Inverness (or reverse, as I did it) through some of the most remote and beautiful landscapes in Scotland.

<div id="route-map-badger" class="not-prose rounded-lg overflow-hidden" style="height:450px;width:100%;"></div>
<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"/>
<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
<script>
(function(){fetch('/api/routes/badger-divide/coords').then(function(r){return r.json()}).then(function(coords){var map=L.map('route-map-badger').setView([55.94,-4.31],8);L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',{attribution:'&copy; OpenStreetMap contributors',maxZoom:18}).addTo(map);var pts=coords.map(function(c){return[c.lat,c.lon]});var poly=L.polyline(pts,{color:'#e74c3c',weight:3}).addTo(map);map.fitBounds(poly.getBounds())})})();
</script>

## Day 1 — Glasgow to Tyndrum

The first day was a long push north through the Lowlands. Gravel tracks, canal paths, and the gradual climb into the Highlands.

### Camp Spot — By a Bothy

Found an open bothy near Crianlarich and sheltered from the rain.

## Day 2 — Tyndrum to Fort William

Over Rannoch Moor and into Glen Coe. One of the most dramatic days of riding I've ever done. The old military road across the moor is rough but the scenery is worth every bump.

## Day 3 — Fort William to Drumnadrochit

Along the Great Glen past Loch Ness. The Caledonian Canal towpath makes for easy cruising.

## Day 4 — Drumnadrochit to Inverness

A short final day rolling into Inverness. Fish and chips by the river to finish.

## Kit

- Bike: Surly Straggler
- Bags: Apidura saddle pack, custom frame bag
- Shelter: MSR Hubba Hubba NX
- Sleep system: Rab Neutrino 400, Therm-a-Rest NeoAir XLite

## Final Thoughts

The Badger Divide is a proper adventure. Well-waymarked, wild camping friendly, and utterly stunning. Highly recommended.
