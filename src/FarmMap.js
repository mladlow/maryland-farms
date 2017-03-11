import React, { Component } from 'react';
import loadjs from 'loadjs';

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
          center: { lat: 40.7413549, lng: -73.9980244 },
          zoom: 8
        });
      },
    });
  }

  render() {
    return <div>
      <div className="map" ref={(mapDiv) => { this.mapRef = mapDiv; }} />
      Hello, world
    </div>;
  }

}

export default FarmMap;
