import React from 'react'
import { injectReducer } from '../../store/reducers'
import CoreLayout from '../../layouts/CoreLayout/Layout';

import container from './containers/SettingContainer'
import reducer from './modules'

function Setting(store, props) {
  injectReducer(store, { key: 'setting', reducer })
  return <CoreLayout component={container} {...props} />
}

export default Setting