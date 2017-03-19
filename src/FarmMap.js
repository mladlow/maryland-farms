import React, { Component } from 'react';
import loadjs from 'loadjs';
import {GOOGLE_API_KEY} from '../data/GoogleApiKey';

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
        const marker = new window.google.maps.Marker({
          position: {lat: 39.327982, lng: -77.1658695},
          map: this.map,
          title: 'A Deck Above Farm',
          animation: window.google.maps.Animation.DROP,
          id: 1,
        });
        const infoWindow = new window.google.maps.InfoWindow({
          content: 'A Deck Above Farm (eventing)',
        });
        marker.addListener('click', () => infoWindow.open(this.map, marker));
      },
    });
  }

  render() {
    return <div className="map" ref={(mapDiv) => { this.mapRef = mapDiv; }} />;
  }

}

export default FarmMap;
