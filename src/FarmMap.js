import React, { Component } from 'react';
import loadjs from 'loadjs';
import {GOOGLE_API_KEY} from '../data/GoogleApiKey';
import stables from '../data/stables.json'

// eslint-disable-next-line
const MARYLAND = {
  lat: 38.801296,
  lng: -78.3894508,
};

const DC = {
  lat: 38.8993277,
  lng: -77.0846065,
};

class FarmMap extends Component {
  constructor(props) {
    super(props);
    this.map = null;
    this.mapRef = null;
  }

  componentDidMount() {
    loadjs(`https://maps.googleapis.com/maps/api/js?key=${GOOGLE_API_KEY}&v=3`, {
      success: () => {
        this.map = new window.google.maps.Map(this.mapRef, {
          center: DC,
          zoom: 8
        });
        const markers = [];
        const infoWindows = [];
        stables.forEach((stable) => {
          const marker = new window.google.maps.Marker({
            position: stable.position,
            map: this.map,
            title: stable.title,
            id: stable.id,
          });
          const infoWindow = new window.google.maps.InfoWindow({
            content: stable.title + "\n" + stable.address,
          });
          marker.addListener('click', () => infoWindow.open(this.map, marker));
          markers.push(marker);
          infoWindows.push(infoWindow);
        });
      },
    });
  }

  render() {
    return <div className="map" ref={(mapDiv) => { this.mapRef = mapDiv; }} />;
  }

}

export default FarmMap;
