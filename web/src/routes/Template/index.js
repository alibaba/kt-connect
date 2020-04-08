import React from 'react';
import { injectReducer } from '../../store/reducers';
import CoreLayout from '../../layouts/CoreLayout/Layout';

import container from './containers/TemplateContainer'
import reducer from './modules'

// 1. change the file name
function Template(store, props) {
  // 2. change the reducer key
  injectReducer(store, { key: 'template', reducer })
  return <CoreLayout component={container} {...props} />
}

export default Template
