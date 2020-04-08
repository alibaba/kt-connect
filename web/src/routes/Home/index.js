import React from 'react';
import { injectReducer } from '../../store/reducers';
import CoreLayout from '../../layouts/CoreLayout/Layout';

import container from './containers/HomeContainer'
import reducer from './modules'

function Home(store, props) {
  injectReducer(store, { key: 'home', reducer })
  return <CoreLayout component={container} {...props} />
}

export default Home
