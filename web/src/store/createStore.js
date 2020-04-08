import { applyMiddleware, compose, createStore as createReduxStore } from 'redux'
import thunk from 'redux-thunk'
import logger from 'redux-logger'
import makeRootReducer from './reducers'

const createStore = (initialState = {}) => {
  // ======================================================
  // Middleware Configuration
  // ======================================================
  const middleware = [thunk, logger]
  // ======================================================
  // Store Instantiation and HMR Setup
  // ======================================================
  const store = createReduxStore(
    makeRootReducer(),
    initialState,
    compose(
      applyMiddleware(...middleware),
    )
  )
  store.asyncReducers = {}

  return store
}

export default createStore