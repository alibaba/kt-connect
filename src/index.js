import React from 'react';
import ReactDOM from 'react-dom';
import createStore from './store/createStore'

import './index.css';
import App from './components/App'
import * as serviceWorker from './serviceWorker'
import createRoutes from './routes'

// Store Initialization
const store = createStore(window.__INITIAL_STATE__)

const routes = createRoutes(store)
ReactDOM.render(<App store={store} routes={routes} />, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
