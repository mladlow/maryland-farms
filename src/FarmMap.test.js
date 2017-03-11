import React from 'react';
import { mount } from 'enzyme';
import FarmMap from './FarmMap';

it('renders without crashing', () => {
  const component = mount(<FarmMap />);
  expect(component.instance().mapRef).not.toBe(null);
});
