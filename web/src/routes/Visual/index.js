import React from 'react';
import { injectReducer } from '../../store/reducers';
import CoreLayout from '../../layouts/CoreLayout/Layout';

import container from './containers/VisualContainer';
import reducer from './modules'

function Visual(store, props) {
  injectReducer(store, { key: 'visual', reducer })
  return <CoreLayout component={container} {...props} />
}

export default Visual
