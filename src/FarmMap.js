import React, { Component } from 'react';
import loadjs from 'loadjs';

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
    loadjs('https://maps.googleapis.com/maps/api/js?key=AIzaSyCSWQCl0L6FeZTMgNqvdiTPWqmTBXoPzV4&v=3', {
      success: () => {
        this.map = new window.google.maps.Map(this.mapRef, {
          center: DC,
          zoom: 8
        });
      },
    });
  }

  render() {
    return <div className="map" ref={(mapDiv) => { this.mapRef = mapDiv; }} />;
  }

}

export default FarmMap;
