"use client"

import { useDriverStreamConnection } from "../hooks/useDriverStreamConnection"
import { MapContainer, Marker, Popup, TileLayer } from 'react-leaflet'
import L from 'leaflet';
import { MapClickHandler } from './MapClickHandler';
import { useMemo, useState, useEffect, useRef } from "react";
import { useGeolocation } from '../hooks/useGeolocation';
import { CarPackageSlug, Coordinate } from "../types";
import { DriverTripOverview } from "./DriverTripOverview";
import * as Geohash from 'ngeohash';
import { RoutingControl } from "./RoutingControl";
import { DriverCard } from "./DriverCard";
import { TripEvents } from "../contracts";

const START_LOCATION: Coordinate = {
  latitude: 28.6139,
  longitude: 77.2090,
}

const driverMarker = new L.Icon({
  iconUrl: "https://www.svgrepo.com/show/25407/car.svg",
  iconSize: [30, 30],
  iconAnchor: [15, 30],
});

const startLocationMarker = new L.Icon({
  iconUrl: "https://www.svgrepo.com/show/535711/user.svg",
  iconSize: [30, 40], // Size of the marker
  iconAnchor: [20, 40], // Anchor point
});

const destinationMarker = new L.Icon({
  iconUrl: "https://upload.wikimedia.org/wikipedia/commons/thumb/e/ed/Map_pin_icon.svg/176px-Map_pin_icon.svg.png",
  iconSize: [40, 40], // Size of the marker
  iconAnchor: [20, 40], // Anchor point
});

export const DriverMap = ({ packageSlug }: { packageSlug: CarPackageSlug }) => {
  const mapRef = useRef<L.Map>(null)
  const [userID, setUserID] = useState<string | null>(null);

  useEffect(() => {
    const storedUserID = localStorage.getItem("driver-id");
    if (storedUserID) {
      setUserID(storedUserID);
    } else {
      const newUserID = crypto.randomUUID();
      localStorage.setItem("driver-id", newUserID);
      setUserID(newUserID);
    }
  }, []);

  // Use geolocation with fallback to New Delhi
  const { latitude, longitude, error: geoError, loading: geoLoading } = useGeolocation({
    watch: true, // Enable real-time location updates
    enableHighAccuracy: true
  });

  const [riderLocation, setRiderLocation] = useState<Coordinate>({
    latitude: latitude ?? START_LOCATION.latitude,
    longitude: longitude ?? START_LOCATION.longitude,
  });

  // Update driver location when geolocation changes
  useEffect(() => {
    if (latitude && longitude) {
      setRiderLocation({ latitude, longitude });
      if (mapRef.current) {
        mapRef.current.setView([latitude, longitude], 13);
      }
    }
  }, [latitude, longitude]);

  const driverGeohash = useMemo(() =>
    Geohash.encode(riderLocation?.latitude, riderLocation?.longitude, 7)
    , [riderLocation?.latitude, riderLocation?.longitude]);

  const {
    error,
    driver,
    tripStatus,
    requestedTrip,
    sendMessage,
    setTripStatus,
    resetTripStatus,
  } = useDriverStreamConnection({
    location: riderLocation,
    geohash: driverGeohash,
    userID,
    packageSlug,
  })

  const handleMapClick = (e: L.LeafletMouseEvent) => {
    setRiderLocation({
      latitude: e.latlng.lat,
      longitude: e.latlng.lng
    })
  }

  const handleAcceptTrip = () => {
    if (!requestedTrip || !requestedTrip.id || !driver) {
      alert("No trip ID found or driver is not set")
      return
    }

    sendMessage({
      type: TripEvents.DriverTripAccept,
      data: {
        tripID: requestedTrip.id,
        riderID: requestedTrip.userID,
        driver: driver,
      }
    })

    setTripStatus(TripEvents.DriverTripAccept)

  }

  const handleDeclineTrip = () => {
    if (!requestedTrip || !requestedTrip.id || !driver) {
      alert("No trip ID found or driver is not set")
      return
    }

    sendMessage({
      type: TripEvents.DriverTripDecline,
      data: {
        tripID: requestedTrip.id,
        riderID: requestedTrip.userID,
        driver: driver,
      }
    })

    setTripStatus(TripEvents.DriverTripDecline)
    resetTripStatus()
  }

  const parsedRoute = useMemo(() =>
    requestedTrip?.route?.geometry[0]?.coordinates
      .map((coord) => [coord?.longitude, coord?.latitude] as [number, number])
    , [requestedTrip])

  // destination is the last coordinate in the route
  const destination = useMemo(() =>
    requestedTrip?.route?.geometry[0]?.coordinates[requestedTrip?.route?.geometry[0]?.coordinates?.length - 1]
    , [requestedTrip])
  // start location is the first coordinate in the route
  const startLocation = useMemo(() =>
    requestedTrip?.route?.geometry[0]?.coordinates[0]
    , [requestedTrip])


  if (error) {
    return <div>Error: {error}</div>
  }

  if (!userID) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg font-semibold">Loading Driver...</div>
      </div>
    );
  }

  return (
    <div className="relative flex flex-col md:flex-row h-screen">
      {/* Location Status Indicator */}
      <div className="absolute top-4 left-4 z-[1000] bg-white rounded-lg shadow-md p-2 text-sm">
        {geoLoading && (
          <div className="flex items-center gap-2 text-blue-600">
            <div className="w-2 h-2 bg-blue-600 rounded-full animate-pulse"></div>
            Getting location...
          </div>
        )}
        {geoError && (
          <div className="flex items-center gap-2 text-orange-600">
            <div className="w-2 h-2 bg-orange-600 rounded-full"></div>
            Using default location
          </div>
        )}
        {latitude && longitude && !geoLoading && (
          <div className="flex items-center gap-2 text-green-600">
            <div className="w-2 h-2 bg-green-600 rounded-full"></div>
            Live location active
          </div>
        )}
      </div>
      <div className="flex-1">
        <MapContainer
          center={[riderLocation.latitude, riderLocation.longitude]}
          zoom={13}
          style={{ height: '100%', width: '100%' }}
          ref={mapRef}
        >
          <TileLayer
            url="https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png"
            attribution="&copy; <a href='https://www.openstreetmap.org/copyright'>OpenStreetMap</a> contributors &copy; <a href='https://carto.com/'>CARTO</a>"
          />

          <Marker
            key={userID}
            position={[riderLocation.latitude, riderLocation.longitude]}
            icon={driverMarker}
          >
            <Popup>
              Driver ID: {userID}
              <br />
              Geohash: {driverGeohash}
              <br />
              {geoError ? 'Using default location' : 'Live location'}
            </Popup>
          </Marker>

          {startLocation && (
            <Marker position={[startLocation.longitude, startLocation.latitude]} icon={startLocationMarker}>
              <Popup>Start Location</Popup>
            </Marker>
          )}

          {destination && (
            <Marker position={[destination.longitude, destination.latitude]} icon={destinationMarker}>
              <Popup>Destination</Popup>
            </Marker>
          )}

          {parsedRoute && (
            <RoutingControl route={parsedRoute} />
          )}

          <MapClickHandler onClick={handleMapClick} />
        </MapContainer>
      </div>

      <div className="flex flex-col md:w-[400px] bg-white border-t md:border-t-0 md:border-l">
        <div className="p-4 border-b">
          <DriverCard driver={driver} packageSlug={packageSlug} />
        </div>
        <div className="flex-1 overflow-y-auto">
          <DriverTripOverview
            trip={requestedTrip}
            status={tripStatus}
            onAcceptTrip={handleAcceptTrip}
            onDeclineTrip={handleDeclineTrip}
          />
        </div>
      </div>
    </div>
  )
}