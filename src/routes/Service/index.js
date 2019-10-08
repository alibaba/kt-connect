import React from 'react';
import { injectReducer } from '../../store/reducers';
import CoreLayout from '../../layouts/CoreLayout/Layout';

import container from './containers/ServiceContainer'
import reducer from './modules'

// 1. change the name
function Service(store, props) {
  // 2. change the reducer key
  injectReducer(store, { key: 'service', reducer })
  return <CoreLayout component={container} {...props} />
}

export default Service
