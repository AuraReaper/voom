'use client';

import Image from 'next/image';
import { useRiderStreamConnection } from '../hooks/useRiderStreamConnection';
import { MapContainer, Marker, Popup, Rectangle, TileLayer } from 'react-leaflet'
import L from 'leaflet';
import { getGeohashBounds } from '../utils/geohash';
import { useMemo, useRef, useState, useEffect } from 'react';
import { MapClickHandler } from './MapClickHandler';
import { Button } from './ui/button';
import { RouteFare, RequestRideProps, TripPreview, HTTPTripStartResponse } from "../types";
import { RoutingControl } from "./RoutingControl";
import { API_URL } from '../constants';
import { RiderTripOverview } from './RiderTripOverview';
import { BackendEndpoints, HTTPTripPreviewRequestPayload, HTTPTripPreviewResponse, HTTPTripStartRequestPayload } from '../contracts';
import { useGeolocation } from '../hooks/useGeolocation';

const userMarker = new L.Icon({
    iconUrl: "https://upload.wikimedia.org/wikipedia/commons/thumb/e/ed/Map_pin_icon.svg/176px-Map_pin_icon.svg.png",
    iconSize: [40, 40], // Size of the marker
    iconAnchor: [20, 40], // Anchor point
});

const currentLocationMarker = new L.Icon({
    iconUrl: "https://www.svgrepo.com/show/535711/user.svg",
    iconSize: [30, 30],
    iconAnchor: [15, 15],
});

const driverMarker = new L.Icon({
    iconUrl: "https://www.svgrepo.com/show/25407/car.svg",
    iconSize: [30, 30],
    iconAnchor: [15, 30],
});

interface RiderMapProps {
    onRouteSelected?: (distance: number) => void;
}

export default function RiderMap({ onRouteSelected }: RiderMapProps) {
    const [trip, setTrip] = useState<TripPreview | null>(null)
    const [selectedCarPackage] = useState<RouteFare | null>(null)
    const [destination, setDestination] = useState<[number, number] | null>(null)
    const mapRef = useRef<L.Map>(null)
    const userID = useMemo(() => crypto.randomUUID(), [])
    const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // Use geolocation with fallback to New Delhi
    const { latitude, longitude, error: geoError, loading: geoLoading } = useGeolocation({ 
        watch: true, // Enable real-time location updates
        enableHighAccuracy: true 
    });

    const location = {
        latitude: latitude ?? 28.6139, // Fallback to New Delhi
        longitude: longitude ?? 77.2090, // Fallback to New Delhi
    };

    // Update map center when location changes
    useEffect(() => {
        if (mapRef.current && latitude && longitude) {
            mapRef.current.setView([latitude, longitude], 13);
        }
    }, [latitude, longitude]);

    const {
        drivers,
        error,
        tripStatus,
        assignedDriver,
        paymentSession,
        resetTripStatus
    } = useRiderStreamConnection(location, userID);

    console.log(tripStatus)

    const handleMapClick = async (e: L.LeafletMouseEvent) => {
        if (trip?.tripID) {
            return
        }

        if (debounceTimeoutRef.current) {
            clearTimeout(debounceTimeoutRef.current);
        }

        debounceTimeoutRef.current = setTimeout(async () => {
            setDestination([e.latlng.lat, e.latlng.lng])

            const data = await requestRidePreview({
                pickup: [location.latitude, location.longitude],
                destination: [e.latlng.lat, e.latlng.lng],
            })
            console.log(data)

            const parsedRoute = data.route.geometry[0].coordinates
                .map((coord) => [coord.longitude, coord.latitude] as [number, number])

            setTrip({
                tripID: "",
                route: parsedRoute,
                rideFare: data.rideFare,
                distance: data.route.distance,
                duration: data.route.duration,
            })

            // Call onRouteSelected with the route distance
            onRouteSelected?.(data.route.distance)
        }, 500);
    }

    const requestRidePreview = async (props: RequestRideProps): Promise<HTTPTripPreviewResponse> => {
        const { pickup, destination } = props
        const payload = {
            userID: userID,
            pickup: {
                latitude: pickup[0],
                longitude: pickup[1],
            },
            destination: {
                latitude: destination[0],
                longitude: destination[1],
            },
        } as HTTPTripPreviewRequestPayload

        const response = await fetch(`${API_URL}${BackendEndpoints.PREVIEW_TRIP}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        })
        const { data } = await response.json() as { data: HTTPTripPreviewResponse }
        return data
    }

    const handleStartTrip = async (fare: RouteFare) => {
        const payload = {
            rideFareID: fare.id,
            userID: userID,
        } as HTTPTripStartRequestPayload

        if (!fare.id) {
            alert("No Fare ID in the payload")
            return
        }

        const response = await fetch(`${API_URL}${BackendEndpoints.START_TRIP}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        })
        const data = await response.json() as HTTPTripStartResponse

        if (response.ok && trip) {
            setTrip((prev) => ({
                ...prev,
                tripID: data.tripID,
            } as TripPreview))

        }

        return data
    }

    const handleCancelTrip = () => {
        setTrip(null)
        setDestination(null)
        resetTripStatus()
    }

    if (error) {
        return <div>Error: {error}</div>
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
            <div className={`${destination ? 'flex-[0.7]' : 'flex-1'}`}>
                <MapContainer
                    center={[location.latitude, location.longitude]}
                    zoom={13}
                    style={{ height: '100%', width: '100%' }}
                    ref={mapRef}
                >
                    <TileLayer
                        url="https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png"
                        attribution="&copy; <a href='https://www.openstreetmap.org/copyright'>OpenStreetMap</a> contributors &copy; <a href='https://carto.com/'>CARTO</a>"
                    />
                    <Marker position={[location.latitude, location.longitude]} icon={currentLocationMarker}>
                        <Popup>
                            Your Current Location
                            {geoError && <><br />Fallback: New Delhi, India</>}
                            {geoLoading && <><br />Getting your location...</>}
                        </Popup>
                    </Marker>

                    {/* Render geohash grid cells */}
                    {drivers?.map((driver) => (
                        <Rectangle
                            key={`grid-${driver?.geohash}`}
                            bounds={getGeohashBounds(driver?.geohash) as L.LatLngBoundsExpression}
                            pathOptions={{
                                color: '#3388ff',
                                weight: 1,
                                fillOpacity: 0.1
                            }}
                        >
                            <Popup>Geohash: {driver?.geohash}</Popup>
                        </Rectangle>
                    ))}

                    {/* Render driver markers */}
                    {drivers?.map((driver) => (
                        <Marker
                            key={driver?.id}
                            position={[driver?.location?.latitude, driver?.location?.longitude]}
                            icon={driverMarker}
                        >
                            <Popup>
                                Driver ID: {driver?.id}
                                <br />
                                Geohash: {driver?.geohash}
                                <br />
                                Name: {driver?.name}
                                <br />
                                Car Plate: {driver?.carPlate}
                                <br />
                                <Image
                                    src={driver?.profilePicture}
                                    alt={`${driver?.name}'s profile picture`}
                                    width={100}
                                    height={100}
                                />
                            </Popup>
                        </Marker>
                    ))}
                    {destination && (
                        <Marker position={destination} icon={userMarker}>
                            <Popup>Destination</Popup>
                        </Marker>
                    )}

                    {selectedCarPackage && (
                        <div className="mt-4 z-[9999] absolute bottom-0 right-0">
                            <Button className="w-full">
                                Request Ride with {selectedCarPackage.packageSlug}
                            </Button>
                        </div>
                    )}
                    {trip && (
                        <RoutingControl route={trip.route} />
                    )}
                    <MapClickHandler onClick={handleMapClick} />
                </MapContainer>
            </div>

            <div className="flex-[0.4]">
                <RiderTripOverview
                    trip={trip}
                    assignedDriver={assignedDriver}
                    status={tripStatus}
                    paymentSession={paymentSession}
                    onPackageSelect={handleStartTrip}
                    onCancel={handleCancelTrip}
                />
            </div>
        </div>
    )
}