import React from 'react';
import { injectReducer } from '../../store/reducers';
import CoreLayout from '../../layouts/CoreLayout/Layout';

import container from './containers/ConnectContainer'
import reducer from './modules'

function Connect(store, props) {
  injectReducer(store, { key: 'connect', reducer })
  return <CoreLayout component={container} {...props} />
}

export default Connect
