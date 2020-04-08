import React from 'react';
import Home from './Home';
import {Redirect} from 'react-router-dom';
import Connect from './Connect';
import Visual from './Visual';
import Setting from './Setting';
import Service from './Service';

export const createRoutes = (store) => {
  return (
    [
      {
        name: 'Dashboard',
        path: '/',
        exact: true,
        render: () => {
          return <Redirect to="/default/visual"/>
        }
      },
      {
        name: 'Dashboard',
        path: '/:namespace/dashboard',
        exact: true,
        render: (props) => {
          return Home(store, props)
        }
      },
      {
        name: 'Connect',
        path: '/:namespace/connect',
        exact: true,
        render: (props) => {
          return Connect(store, props)
        }
      },
      {
        name: 'Visual',
        path: '/:namespace/visual',
        exact: true,
        render: (props) => {
          return Visual(store, props)
        }
      },
      {
        name: 'Service',
        path: '/:namespace/visual/:name',
        exact: true,
        render: (props) => {
          return Service(store, props)
        }
      },
      {
        name: 'Setting',
        path: '/setting',
        render: (props) => {
          return Setting(store, props)
        }
      }
    ])
}

export default createRoutes;
